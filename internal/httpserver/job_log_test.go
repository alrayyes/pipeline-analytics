package httpserver_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

var errStubReader = errors.New("reader exploded")

type stubLogReader struct {
	tail ingestion.JobLogTail
	err  error
	got  ingestion.JobLogRequest
}

func (s *stubLogReader) JobLogTail(_ context.Context, req ingestion.JobLogRequest) (ingestion.JobLogTail, error) {
	s.got = req

	return s.tail, s.err
}

// logTestServer seeds one run with one job and returns the server, the run's
// id and the job's id as the steps endpoint reports it.
func logTestServer(t *testing.T, reader ingestion.JobLogReader) (srv testServer, runID, jobID string) {
	t.Helper()

	srv = newTestServerWithDeps(t, func(deps *httpserver.Deps) {
		store, ok := deps.RunStore.(*ingestionsqlite.Store)
		require.True(t, ok)

		readers := map[ingestion.Forge]ingestion.JobLogReader{}
		if reader != nil {
			readers[ingestion.ForgeGitHub] = reader
		}

		deps.JobLogs = ingestion.NewJobLogService(store, store, readers)
	})

	repoID := seedRepo(t, srv)
	run := seedRun(t, srv, repoID, "1", 0, 5)
	seedJobStep(t, srv, run.ID, "100", "go test", "failure")

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/runs/"+run.ID+"/steps", nil)))
	require.Equal(t, http.StatusOK, rec.Code)

	var detail struct {
		Steps []struct {
			JobID string `json:"jobId"`
		} `json:"steps"`
	}

	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &detail))
	require.NotEmpty(t, detail.Steps)

	return srv, run.ID, detail.Steps[0].JobID
}

func getLog(srv testServer, runID, jobID, query string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, srv.authenticated(httptest.NewRequest(http.MethodGet, "/api/runs/"+runID+"/jobs/"+jobID+"/log"+query, nil)))

	return rec
}

func TestJobLog(t *testing.T) {
	t.Parallel()

	t.Run("returns the log tail from the forge, ANSI untouched", func(t *testing.T) {
		t.Parallel()

		reader := &stubLogReader{tail: ingestion.JobLogTail{Lines: []string{"ok", "\x1b[31mFAIL\x1b[0m <script>alert(1)</script>"}, Truncated: true}}
		srv, runID, jobID := logTestServer(t, reader)

		rec := getLog(srv, runID, jobID, "?lines=50")
		require.Equal(t, http.StatusOK, rec.Code)

		var got map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, true, got["available"])
		require.Equal(t, true, got["truncated"])
		require.Equal(t, []any{"ok", "\x1b[31mFAIL\x1b[0m <script>alert(1)</script>"}, got["lines"])
		require.Contains(t, got["forgeUrl"], "/job/100")
		require.Equal(t, "100", reader.got.ForgeJobID, "the forge is asked for its own job id, not ours")
		require.Equal(t, 50, reader.got.Lines)
		require.Equal(t, "ghp_supersecrettoken1234", reader.got.Token)
	})

	t.Run("a log the forge can't give is a 200 with a reason and the deep link", func(t *testing.T) {
		t.Parallel()

		srv, runID, jobID := logTestServer(t, &stubLogReader{err: ingestion.ErrLogExpired})

		rec := getLog(srv, runID, jobID, "")
		require.Equal(t, http.StatusOK, rec.Code)

		var got map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Equal(t, false, got["available"])
		require.Equal(t, "expired", got["reason"])
		require.Equal(t, []any{}, got["lines"])
		require.Contains(t, got["forgeUrl"], "/job/100")
	})

	t.Run("a forge with no log reader is unsupported", func(t *testing.T) {
		t.Parallel()

		srv, runID, jobID := logTestServer(t, nil)

		var got map[string]any
		require.NoError(t, json.Unmarshal(getLog(srv, runID, jobID, "").Body.Bytes(), &got))
		require.Equal(t, "unsupported", got["reason"])
	})

	t.Run("an unexpected reader failure is a 500", func(t *testing.T) {
		t.Parallel()

		srv, runID, jobID := logTestServer(t, &stubLogReader{err: errStubReader})

		require.Equal(t, http.StatusInternalServerError, getLog(srv, runID, jobID, "").Code)
	})

	t.Run("an unknown job, or one from another run, is not found", func(t *testing.T) {
		t.Parallel()

		srv, runID, jobID := logTestServer(t, &stubLogReader{})

		require.Equal(t, http.StatusNotFound, getLog(srv, runID, "nope", "").Code)
		require.Equal(t, http.StatusNotFound, getLog(srv, "other-run", jobID, "").Code)
	})

	t.Run("lines outside 1 to 1000, or not a number, are rejected", func(t *testing.T) {
		t.Parallel()

		srv, runID, jobID := logTestServer(t, &stubLogReader{})

		for _, q := range []string{"?lines=0", "?lines=1001", "?lines=-3", "?lines=abc"} {
			require.Equal(t, http.StatusBadRequest, getLog(srv, runID, jobID, q).Code, q)
		}
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		srv, runID, jobID := logTestServer(t, &stubLogReader{})

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/runs/"+runID+"/jobs/"+jobID+"/log", nil))
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestMCPGetJobLog(t *testing.T) {
	t.Parallel()

	t.Run("returns the log tail, as the REST endpoint does", func(t *testing.T) {
		t.Parallel()

		reader := &stubLogReader{tail: ingestion.JobLogTail{Lines: []string{"one", "\x1b[31mtwo\x1b[0m"}}}
		srv, runID, jobID := logTestServer(t, reader)

		got := callTool[struct {
			Available bool     `json:"available"`
			Lines     []string `json:"lines"`
			ForgeURL  string   `json:"forgeUrl"`
		}](t, connectMCP(t, srv), "get_job_log", map[string]any{"runId": runID, "jobId": jobID, "lines": 20})

		require.True(t, got.Available)
		require.Equal(t, []string{"one", "\x1b[31mtwo\x1b[0m"}, got.Lines)
		require.Contains(t, got.ForgeURL, "/job/100")
		require.Equal(t, 20, reader.got.Lines)
	})

	t.Run("a log the forge can't give carries the reason", func(t *testing.T) {
		t.Parallel()

		srv, runID, jobID := logTestServer(t, &stubLogReader{err: ingestion.ErrLogForbidden})

		got := callTool[struct {
			Available bool   `json:"available"`
			Reason    string `json:"reason"`
		}](t, connectMCP(t, srv), "get_job_log", map[string]any{"runId": runID, "jobId": jobID})

		require.False(t, got.Available)
		require.Equal(t, "forbidden", got.Reason)
	})

	t.Run("an unknown job is a tool error", func(t *testing.T) {
		t.Parallel()

		srv, runID, _ := logTestServer(t, &stubLogReader{})

		res, err := connectMCP(t, srv).CallTool(context.Background(), &mcp.CallToolParams{Name: "get_job_log", Arguments: map[string]any{"runId": runID, "jobId": "nope"}})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})

	t.Run("lines outside 1 to 1000 is a tool error", func(t *testing.T) {
		t.Parallel()

		srv, runID, jobID := logTestServer(t, &stubLogReader{})

		res, err := connectMCP(t, srv).CallTool(context.Background(), &mcp.CallToolParams{Name: "get_job_log", Arguments: map[string]any{"runId": runID, "jobId": jobID, "lines": 5000}})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})

	t.Run("a build without a log service says so", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		res, err := connectMCP(t, srv).CallTool(context.Background(), &mcp.CallToolParams{Name: "get_job_log", Arguments: map[string]any{"runId": "r", "jobId": "j"}})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})

	t.Run("an unexpected reader failure is a tool error", func(t *testing.T) {
		t.Parallel()

		srv, runID, jobID := logTestServer(t, &stubLogReader{err: errStubReader})

		res, err := connectMCP(t, srv).CallTool(context.Background(), &mcp.CallToolParams{Name: "get_job_log", Arguments: map[string]any{"runId": runID, "jobId": jobID}})
		require.NoError(t, err)
		require.True(t, res.IsError)
	})
}
