package httpserver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	authsqlite "github.com/alrayyes/pipeline-analytics/internal/auth/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	metricssqlite "github.com/alrayyes/pipeline-analytics/internal/metrics/sqlite"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// newTestServerWithAuth is newTestServer, but Deps.Auth is also wired,
// the same way newAuthTestServer (auth_test.go) wires it -- needed only
// by a test that issues an API token through the real HTTP endpoint;
// newTestServer leaves Deps.Auth nil since nothing else in this package
// calls it.
func newTestServerWithAuth(t *testing.T) testServer {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	ingestionStore := ingestionsqlite.NewStore(conn, make([]byte, 32))

	authStore := authsqlite.NewStore(conn)
	web, err := gowebauthn.New(&gowebauthn.Config{
		RPID:          authTestRPID,
		RPDisplayName: "pipeline-analytics",
		RPOrigins:     []string{authTestOrigin},
	})
	require.NoError(t, err)

	ctx := context.Background()
	user, err := authStore.CreateUser(ctx, []byte("test-handle"), "admin")
	require.NoError(t, err)
	sessionID, err := authStore.CreateSession(ctx, user.ID)
	require.NoError(t, err)

	handler := httpserver.New(httpserver.Deps{
		IngestionStore: ingestionStore,
		RunStore:       ingestionStore,
		Metrics:        metrics.NewService(metricssqlite.NewStore(conn)),
		Auth:           auth.NewService(web, authStore),
		AuthStore:      authStore,
		Version:        "test-version",
		Assets:         testAssets,
	})

	cookie := &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	return testServer{Handler: handler, sessionCookie: cookie, runStore: ingestionStore}
}

// authRoundTripper attaches srv's session cookie, or a bearer token, to
// every outgoing request -- what an mcp.StreamableClientTransport needs
// to reach a gated endpoint, since it takes a plain *http.Client rather
// than a cookie jar or header map.
type authRoundTripper struct {
	cookie *http.Cookie
	bearer string
}

func (rt authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())

	if rt.cookie != nil {
		req.AddCookie(rt.cookie)
	}

	if rt.bearer != "" {
		req.Header.Set("Authorization", "Bearer "+rt.bearer)
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("round trip: %w", err)
	}

	return resp, nil
}

// connectMCP starts a real network server for srv (the MCP client
// transport does real HTTP requests, not httptest.NewRecorder round
// trips), connects an MCP client to POST /api/mcp using srv's session
// cookie, and registers cleanup.
func connectMCP(t *testing.T, srv testServer) *mcp.ClientSession {
	t.Helper()

	httpSrv := httptest.NewServer(srv)
	t.Cleanup(httpSrv.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             httpSrv.URL + "/api/mcp",
		HTTPClient:           &http.Client{Transport: authRoundTripper{cookie: srv.sessionCookie}},
		DisableStandaloneSSE: true,
	}

	session, err := client.Connect(context.Background(), transport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	return session
}

// callTool calls name with args, requires it succeeded, and decodes its
// structured content into T.
func callTool[T any](t *testing.T, session *mcp.ClientSession, name string, args map[string]any) T {
	t.Helper()

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	require.NoError(t, err)
	require.False(t, res.IsError, "tool call reported an error: %+v", res.Content)

	raw, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)

	var out T
	require.NoError(t, json.Unmarshal(raw, &out))

	return out
}

func TestMCPEndpointRequiresAuth(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/mcp", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMCPToolsList(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	session := connectMCP(t, srv)

	res, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)

	names := make([]string, 0, len(res.Tools))
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}

	require.ElementsMatch(t, []string{
		"list_pipelines",
		"get_pipeline",
		"list_pipeline_steps",
		"list_pipeline_flaky_runs",
		"get_repo_usage",
		"list_unhealthy_steps",
	}, names, "tool list must be exactly the six read tools, no write tool")
}

func TestMCPListPipelines(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	repoID := seedRepo(t, srv)
	seedRun(t, srv, repoID, "1", 0, 5)
	seedRun(t, srv, repoID, "2", 10, 5)

	session := connectMCP(t, srv)

	got := callTool[pipelineListResponse](t, session, "list_pipelines", nil)

	require.False(t, got.HasMore)
	require.Len(t, got.Pipelines, 1)
	require.Equal(t, seedRunPipelineName, got.Pipelines[0]["name"])
	require.Equal(t, "healthy", got.Pipelines[0]["healthStatus"])
	require.Equal(t, repoID, got.Pipelines[0]["repoId"])
}

func TestMCPGetPipeline(t *testing.T) {
	t.Parallel()

	t.Run("returns trend data for a tracked pipeline", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		seedRun(t, srv, repoID, "1", 0, 5)

		session := connectMCP(t, srv)

		list := callTool[pipelineListResponse](t, session, "list_pipelines", nil)
		require.Len(t, list.Pipelines, 1)
		pipelineID, _ := list.Pipelines[0]["id"].(string)
		require.NotEmpty(t, pipelineID)

		got := callTool[map[string]any](t, session, "get_pipeline", map[string]any{"pipelineId": pipelineID})

		require.Equal(t, seedRunPipelineName, got["name"])
		require.Contains(t, got, "durationTrend")
		require.Contains(t, got, "failureRateTrend")
	})

	t.Run("unknown pipeline id is a tool error, not a crash", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		session := connectMCP(t, srv)

		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "get_pipeline",
			Arguments: map[string]any{"pipelineId": "not-a-real-id"},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})
}

func TestMCPListPipelineSteps(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	repoID := seedRepo(t, srv)
	run := seedRun(t, srv, repoID, "1", 0, 30)
	seedJobStep(t, srv, run.ID, "100", "build", "success")

	session := connectMCP(t, srv)

	list := callTool[pipelineListResponse](t, session, "list_pipelines", nil)
	require.Len(t, list.Pipelines, 1)
	pipelineID, _ := list.Pipelines[0]["id"].(string)

	got := callTool[struct {
		Steps []map[string]any `json:"steps"`
	}](t, session, "list_pipeline_steps", map[string]any{"pipelineId": pipelineID})

	require.Len(t, got.Steps, 1)
	require.Equal(t, "build", got.Steps[0]["name"])
}

func TestMCPListPipelineFlakyRuns(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	repoID := seedRepo(t, srv)
	run := seedRun(t, srv, repoID, "1", 0, 30)
	seedJobStep(t, srv, run.ID, "100", "flaky-step", "failure")

	session := connectMCP(t, srv)

	list := callTool[pipelineListResponse](t, session, "list_pipelines", nil)
	require.Len(t, list.Pipelines, 1)
	pipelineID, _ := list.Pipelines[0]["id"].(string)

	got := callTool[struct {
		Runs []map[string]any `json:"runs"`
	}](t, session, "list_pipeline_flaky_runs", map[string]any{
		"pipelineId": pipelineID,
		"step":       "flaky-step",
	})

	require.Len(t, got.Runs, 1)
	require.Equal(t, run.ID, got.Runs[0]["runId"])
}

func TestMCPGetRepoUsage(t *testing.T) {
	t.Parallel()

	t.Run("returns usage for a tracked repo", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)
		run := seedRun(t, srv, repoID, "1", 0, 30)
		seedJobStep(t, srv, run.ID, "100", "build", "success")

		session := connectMCP(t, srv)

		got := callTool[struct {
			Usage []map[string]any `json:"usage"`
		}](t, session, "get_repo_usage", map[string]any{"repoId": repoID})

		require.Len(t, got.Usage, 1)
		require.Equal(t, seedRunPipelineName, got.Usage[0]["workflow"])
	})

	t.Run("unknown repo id is a tool error", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		session := connectMCP(t, srv)

		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "get_repo_usage",
			Arguments: map[string]any{"repoId": "not-a-real-repo"},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})
}

func TestMCPListUnhealthySteps(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	repoID := seedRepo(t, srv)
	run := seedRun(t, srv, repoID, "1", 0, 30)

	job, err := srv.runStore.UpsertJob(context.Background(), ingestion.Job{
		RunID:       run.ID,
		ForgeJobID:  "100",
		Name:        "build",
		Status:      "completed",
		Conclusion:  "failure",
		QueuedAt:    at(0),
		StartedAt:   at(0),
		CompletedAt: at(30),
		ForgeURL:    "https://github.com/alrayyes/pipeline-analytics/actions/runs/1/job/100",
	})
	require.NoError(t, err)

	err = srv.runStore.ReplaceSteps(context.Background(), job.ID, []ingestion.Step{
		{Number: 1, Name: "checkout", Status: "completed", Conclusion: "success", StartedAt: at(0), CompletedAt: at(1)},
		{Number: 2, Name: "deploy", Status: "completed", Conclusion: "failure", StartedAt: at(1), CompletedAt: at(20)},
	})
	require.NoError(t, err)

	session := connectMCP(t, srv)

	got := callTool[unhealthyStepsResponse](t, session, "list_unhealthy_steps", nil)

	require.Len(t, got.Groups, 1)
	require.Equal(t, seedRunPipelineName, got.Groups[0]["pipelineName"])
	require.Equal(t, repoID, got.Groups[0]["repoId"])

	steps, _ := got.Groups[0]["steps"].([]any)
	require.Len(t, steps, 1)
	step, _ := steps[0].(map[string]any)
	require.Equal(t, "deploy", step["name"])
}

func TestMCPBearerTokenAuthenticates(t *testing.T) {
	t.Parallel()

	srv := newTestServerWithAuth(t)

	issueReq := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/auth/tokens", nil))
	issueRec := httptest.NewRecorder()
	srv.ServeHTTP(issueRec, issueReq)
	require.Equal(t, http.StatusCreated, issueRec.Code)

	var issued struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(issueRec.Body.Bytes(), &issued))
	require.NotEmpty(t, issued.Token)

	httpSrv := httptest.NewServer(srv)
	t.Cleanup(httpSrv.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             httpSrv.URL + "/api/mcp",
		HTTPClient:           &http.Client{Transport: authRoundTripper{bearer: issued.Token}},
		DisableStandaloneSSE: true,
	}

	session, err := client.Connect(context.Background(), transport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)
	require.NotEmpty(t, res.Tools)
}
