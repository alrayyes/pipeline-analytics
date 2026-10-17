package httpserver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	authsqlite "github.com/alrayyes/pipeline-analytics/internal/auth/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	metricssqlite "github.com/alrayyes/pipeline-analytics/internal/metrics/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/settings"
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

// connectMCPWithBearer is connectMCP for a caller that holds an API token and
// no session, which is how an agent outside the dashboard reaches the server.
func connectMCPWithBearer(t *testing.T, srv testServer, token string) *mcp.ClientSession {
	t.Helper()

	httpSrv := httptest.NewServer(srv)
	t.Cleanup(httpSrv.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             httpSrv.URL + "/api/mcp",
		HTTPClient:           &http.Client{Transport: authRoundTripper{bearer: token}},
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
		"get_failure_insights",
		"list_runs",
		"get_run_steps",
		"get_job_log",
		"list_flaky_steps",
		"list_branches",
		"get_version",
		"list_repos",
		"get_repo",
		"list_repo_identifiers",
		"get_github_rate_limit",
		"get_settings",
		"update_settings",
		"rerun_run",
		"cancel_run",
		"quarantine_step",
		"unquarantine_step",
		"register_repo",
		"discover_repos",
		"untrack_repo",
	}, names, "tool list must be exactly the declared tools")
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

func TestMCPListPipelines_HealthFilterAndSort(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	repoID := seedRepo(t, srv)
	seedRun(t, srv, repoID, "ok", 0, 5) // "CI", healthy
	seedFailingRuns(t, srv, repoID, "Broken", 100)

	session := connectMCP(t, srv)

	t.Run("the health filter reaches the tool, across pages", func(t *testing.T) {
		t.Parallel()

		got := callTool[pipelineListResponse](t, session, "list_pipelines", map[string]any{"health": "unhealthy", "limit": 1})

		require.Len(t, got.Pipelines, 1)
		require.Equal(t, "Broken", got.Pipelines[0]["name"])
		require.False(t, got.HasMore)
	})

	t.Run("the sort reaches the tool", func(t *testing.T) {
		t.Parallel()

		got := callTool[pipelineListResponse](t, session, "list_pipelines", map[string]any{"sort": "lastRun"})

		require.Equal(t, []string{"Broken", "CI"}, pipelineNames(got))
	})

	t.Run("an unknown health value is a tool error, not an empty list", func(t *testing.T) {
		t.Parallel()

		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "list_pipelines",
			Arguments: map[string]any{"health": "broken"},
		})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})
}

func TestMCPListRuns(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	repoID := seedRepo(t, srv)
	failed := seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "1", Status: "completed", Conclusion: "failure", StartedAt: at(10), CompletedAt: at(15)})
	seedRunWith(t, srv, repoID, ingestion.Run{ForgeRunID: "2", Status: "completed", Conclusion: "success", StartedAt: at(20), CompletedAt: at(25)})
	seedJobStep(t, srv, failed.ID, "10", "deploy", "timed_out")

	session := connectMCP(t, srv)

	type runList struct {
		Runs []struct {
			ID      string `json:"id"`
			Outcome string `json:"outcome"`
			Steps   []struct {
				Name    string `json:"name"`
				Outcome string `json:"outcome"`
			} `json:"steps"`
		} `json:"runs"`
		HasMore bool `json:"hasMore"`
	}

	all := callTool[runList](t, session, "list_runs", nil)
	require.Len(t, all.Runs, 2)
	require.Equal(t, "passed", all.Runs[0].Outcome, "newest first")

	onlyFailed := callTool[runList](t, session, "list_runs", map[string]any{"status": "failed"})
	require.Len(t, onlyFailed.Runs, 1)
	require.Equal(t, failed.ID, onlyFailed.Runs[0].ID)
	require.Equal(t, "failed", onlyFailed.Runs[0].Steps[0].Outcome, "a timed_out step is failed")

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_runs", Arguments: map[string]any{"status": "bogus"}})
	require.NoError(t, err)
	require.True(t, res.IsError, "an unknown status is a tool error, not an empty list")
}

func TestMCPGetRunSteps(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	run := seedRunWith(t, srv, seedRepo(t, srv), ingestion.Run{ForgeRunID: "1", Status: "completed", Conclusion: "failure", StartedAt: at(10), CompletedAt: at(15)})
	seedJobStep(t, srv, run.ID, "10", "build", "success")
	seedJobStep(t, srv, run.ID, "11", "deploy", "failure")

	session := connectMCP(t, srv)

	got := callTool[struct {
		RunID string `json:"runId"`
		Steps []struct {
			Name    string `json:"name"`
			Outcome string `json:"outcome"`
		} `json:"steps"`
	}](t, session, "get_run_steps", map[string]any{"runId": run.ID})

	require.Equal(t, run.ID, got.RunID)
	require.Len(t, got.Steps, 2)
	require.Equal(t, "failed", got.Steps[1].Outcome)

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "get_run_steps", Arguments: map[string]any{"runId": "nope"}})
	require.NoError(t, err)
	require.True(t, res.IsError, "an unknown run is a tool error")
}

func TestMCPGetFailureInsights(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	repoID := seedRepo(t, srv)
	failed := seedRecentRun(t, srv, repoID, seedRunPipelineName, "1", "failure", time.Hour)
	seedRecentRun(t, srv, repoID, seedRunPipelineName, "2", "success", 2*time.Hour)
	seedJobStep(t, srv, failed.ID, "10", "deploy", "timed_out")

	session := connectMCP(t, srv)

	got := callTool[struct {
		Window     string `json:"window"`
		TotalRuns  int    `json:"totalRuns"`
		FailedRuns int    `json:"failedRuns"`
	}](t, session, "get_failure_insights", map[string]any{"window": "30d"})

	require.Equal(t, "30d", got.Window)
	require.Equal(t, 2, got.TotalRuns)
	require.Equal(t, 1, got.FailedRuns)

	def := callTool[struct {
		Window string `json:"window"`
	}](t, session, "get_failure_insights", nil)
	require.Equal(t, "7d", def.Window, "an omitted window reports the server's default")
}

func TestMCPListFlakySteps(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	seedFlakyHistory(t, srv, seedRepo(t, srv), "success", "failure", "success")

	session := connectMCP(t, srv)

	got := callTool[flakyStepsBody](t, session, "list_flaky_steps", map[string]any{"window": "24h"})

	require.Equal(t, "24h", got.Window)
	require.Len(t, got.Steps, 1)
	require.Equal(t, "test", got.Steps[0].Name)
	require.Equal(t, 3, got.Steps[0].RunCount)
	require.Equal(t, []string{"passed", "failed", "passed"}, got.Steps[0].RecentOutcomes)
}

func TestMCPGetVersion(t *testing.T) {
	t.Parallel()

	session := connectMCP(t, newTestServer(t, nil))

	got := callTool[map[string]any](t, session, "get_version", nil)

	require.Equal(t, "test-version", got["version"])
}

func TestMCPListRepos(t *testing.T) {
	t.Parallel()

	type repoList struct {
		Repos   []map[string]any `json:"repos"`
		HasMore bool             `json:"hasMore"`
	}

	srv := newTestServer(t, nil)
	firstID := seedRepoWithToken(t, srv, "alrayyes/first", "ghp_supersecrettoken1234")
	secondID := seedRepoWithToken(t, srv, "alrayyes/second", "ghp_anothersecret5678")
	session := connectMCP(t, srv)

	t.Run("lists the tracked repos with the token masked", func(t *testing.T) {
		t.Parallel()

		got := callTool[repoList](t, session, "list_repos", nil)

		require.Len(t, got.Repos, 2)
		require.Equal(t, firstID, got.Repos[0]["id"])
		require.NotContains(t, got.Repos[0]["tokenMasked"], "supersecret")
	})

	t.Run("a forge filter keeps that forge's repos", func(t *testing.T) {
		t.Parallel()

		got := callTool[repoList](t, session, "list_repos", map[string]any{"forge": "github"})

		require.Len(t, got.Repos, 2)
	})

	t.Run("a forge filter that matches nothing returns no repos", func(t *testing.T) {
		t.Parallel()

		got := callTool[repoList](t, session, "list_repos", map[string]any{"forge": "forgejo"})

		require.Empty(t, got.Repos)
	})

	t.Run("limit pages the list and reports that more follow", func(t *testing.T) {
		t.Parallel()

		got := callTool[repoList](t, session, "list_repos", map[string]any{"limit": 1})

		require.Len(t, got.Repos, 1)
		require.True(t, got.HasMore)
	})

	t.Run("offset skips to the next repo", func(t *testing.T) {
		t.Parallel()

		got := callTool[repoList](t, session, "list_repos", map[string]any{"limit": 1, "offset": 1})

		require.Len(t, got.Repos, 1)
		require.Equal(t, secondID, got.Repos[0]["id"])
		require.False(t, got.HasMore)
	})
}

func TestMCPGetRepo(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	id := seedRepoWithToken(t, srv, "alrayyes/first", "ghp_supersecrettoken1234")
	session := connectMCP(t, srv)

	t.Run("returns the repo with the token masked", func(t *testing.T) {
		t.Parallel()

		got := callTool[map[string]any](t, session, "get_repo", map[string]any{"repoId": id})

		require.Equal(t, "alrayyes/first", got["identifier"])
		require.Equal(t, "****1234", got["tokenMasked"])
	})

	t.Run("an unknown repo is not found", func(t *testing.T) {
		t.Parallel()

		require.Contains(t, toolFailure(t, session, "get_repo", map[string]any{"repoId": "no-such-repo"}), "not found")
	})

	t.Run("a store failure is an error, not a not-found", func(t *testing.T) {
		t.Parallel()

		failing := newTestServerWithStore(t, nil, testAssets, func(s ingestion.Store) ingestion.Store {
			return failingGetStore{Store: s}
		})
		failingSession := connectMCP(t, failing)

		message := toolFailure(t, failingSession, "get_repo", map[string]any{"repoId": "any-id"})

		require.NotContains(t, message, "not found")
	})
}

func TestMCPListRepoIdentifiers(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)
	seedRepo(t, srv)
	session := connectMCP(t, srv)

	t.Run("lists a forge's tracked identifiers", func(t *testing.T) {
		t.Parallel()

		got := callTool[struct {
			Identifiers []string `json:"identifiers"`
		}](t, session, "list_repo_identifiers", map[string]any{"forge": "github"})

		require.Equal(t, []string{"alrayyes/pipeline-analytics"}, got.Identifiers)
	})

	t.Run("no forge at all is refused by the tool's schema", func(t *testing.T) {
		t.Parallel()

		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_repo_identifiers", Arguments: map[string]any{}})

		require.NoError(t, err)
		require.True(t, res.IsError)
		require.Contains(t, toolErrorText(t, res), "forge")
	})

	// Both of these pass the schema, so only the handler's own check, shared
	// with the endpoint, refuses them.
	for _, forge := range []string{"", "gitlab"} {
		t.Run("a forge of "+forge+" is refused as on the endpoint", func(t *testing.T) {
			t.Parallel()

			res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_repo_identifiers", Arguments: map[string]any{"forge": forge}})

			require.NoError(t, err)
			require.True(t, res.IsError)
			require.Contains(t, toolErrorText(t, res), "forge is required")
		})
	}
}

// toolErrorText is the message of a tool call that reported an error.
func toolErrorText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()

	require.NotEmpty(t, res.Content)

	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "error content is %T, want text", res.Content[0])

	return text.Text
}

func TestMCPGetGitHubRateLimit(t *testing.T) {
	t.Parallel()

	type tokenUsage struct {
		TokenMasked string         `json:"tokenMasked"`
		Repos       []string       `json:"repos"`
		Status      map[string]any `json:"status"`
	}

	t.Run("groups the repos that share a token, with its status", func(t *testing.T) {
		t.Parallel()

		resetAt := time.Unix(1700000000, 0).UTC()
		srv := newTestServerWithRateLimits(t, map[string]ingestion.RateLimitSnapshot{
			"ghp_shared": {Limit: 5000, Remaining: 4900, Used: 100, Resource: "core", ResetAt: resetAt},
		})
		seedRepoWithToken(t, srv, "alrayyes/repo-one", "ghp_shared")
		seedRepoWithToken(t, srv, "alrayyes/repo-two", "ghp_shared")
		session := connectMCP(t, srv)

		got := callTool[struct {
			Tokens []tokenUsage `json:"tokens"`
		}](t, session, "get_github_rate_limit", nil)

		require.Len(t, got.Tokens, 1)
		require.ElementsMatch(t, []string{"alrayyes/repo-one", "alrayyes/repo-two"}, got.Tokens[0].Repos)
		require.InDelta(t, 4900, got.Tokens[0].Status["remaining"], 0)
		require.NotContains(t, got.Tokens[0].TokenMasked, "shared")
	})

	t.Run("a token never seen on a request has no status yet", func(t *testing.T) {
		t.Parallel()

		srv := newTestServerWithRateLimits(t, nil)
		seedRepoWithToken(t, srv, "alrayyes/repo-one", "ghp_unseen")
		session := connectMCP(t, srv)

		got := callTool[struct {
			Tokens []tokenUsage `json:"tokens"`
		}](t, session, "get_github_rate_limit", nil)

		require.Len(t, got.Tokens, 1)
		require.Nil(t, got.Tokens[0].Status)
	})
}

func TestMCPGetSettings(t *testing.T) {
	t.Parallel()

	t.Run("returns the account's settings beside the defaults", func(t *testing.T) {
		t.Parallel()

		session := connectMCP(t, newSettingsTestServer(t))

		got := callTool[map[string]any](t, session, "get_settings", nil)

		require.Equal(t, settings.DefaultTheme, got["theme"])
		require.Equal(t, settings.DefaultTelemetryWindow, got["telemetryWindow"])

		defaults, _ := got["defaults"].(map[string]any)
		require.Equal(t, settings.DefaultTheme, defaults["theme"])
	})

	t.Run("an API token reads them too, as on the endpoint", func(t *testing.T) {
		t.Parallel()

		srv, authStore := newSettingsTestServerWithAuthStore(t)

		userID, err := authStore.Session(context.Background(), srv.sessionCookie.Value)
		require.NoError(t, err)

		_, raw, err := authStore.CreateToken(context.Background(), userID, 0)
		require.NoError(t, err)

		got := callTool[map[string]any](t, connectMCPWithBearer(t, srv, raw), "get_settings", nil)

		require.Equal(t, settings.DefaultTheme, got["theme"])
	})

	t.Run("a build with no settings service reports it instead of failing", func(t *testing.T) {
		t.Parallel()

		res, err := connectMCP(t, newTestServer(t, nil)).CallTool(context.Background(), &mcp.CallToolParams{Name: "get_settings"})

		require.NoError(t, err)
		require.True(t, res.IsError)
		require.Contains(t, toolErrorText(t, res), "settings are not available")
	})
}
