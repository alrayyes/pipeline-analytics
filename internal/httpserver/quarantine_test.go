package httpserver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	"github.com/stretchr/testify/require"
)

type quarantineBody struct {
	Quarantined bool `json:"quarantined"`
	Quarantine  *struct {
		Note          string    `json:"note"`
		QuarantinedAt time.Time `json:"quarantinedAt"`
		ExpiresAt     time.Time `json:"expiresAt"`
	} `json:"quarantine"`
}

type quarantinedFlakyList struct {
	Steps []struct {
		PipelineID  string         `json:"pipelineId"`
		Name        string         `json:"name"`
		FlakeRate   float64        `json:"flakeRate"`
		Quarantined bool           `json:"quarantined"`
		Quarantine  map[string]any `json:"quarantine"`
	} `json:"steps"`
}

func quarantinePath(pipelineID, step string) string {
	return "/api/pipelines/" + pipelineID + "/steps/" + url.PathEscape(step) + "/quarantine"
}

// flakyServer seeds one flaky "test" step in the "CI" pipeline and returns the
// server, its pipeline id and its auth store.
func flakyServer(t *testing.T) (testServer, string, auth.Store) {
	t.Helper()

	var authStore auth.Store

	srv := newTestServerWithDeps(t, func(deps *httpserver.Deps) { authStore = deps.AuthStore })
	seedFlakyHistory(t, srv, seedRepo(t, srv), "success", "failure", "success", "success")

	_, body := getFlakySteps(t, srv, "")
	require.Len(t, body.Steps, 1)

	return srv, body.Steps[0].PipelineID, authStore
}

func listQuarantined(t *testing.T, srv testServer) quarantinedFlakyList {
	t.Helper()

	rec := do(srv, http.MethodGet, "/api/steps/flaky", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var out quarantinedFlakyList
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))

	return out
}

func TestQuarantine(t *testing.T) {
	t.Parallel()

	t.Run("quarantining a step marks it for 30 days and the list reports it", func(t *testing.T) {
		t.Parallel()

		srv, id, _ := flakyServer(t)

		rec := do(srv, http.MethodPut, quarantinePath(id, "test"), `{"note":"waits on the shared database"}`)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var got quarantineBody
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.True(t, got.Quarantined)
		require.NotNil(t, got.Quarantine)
		require.Equal(t, "waits on the shared database", got.Quarantine.Note)
		require.WithinDuration(t, time.Now(), got.Quarantine.QuarantinedAt, time.Minute)
		require.WithinDuration(t, time.Now().Add(30*24*time.Hour), got.Quarantine.ExpiresAt, time.Minute)

		list := listQuarantined(t, srv)
		require.Len(t, list.Steps, 1, "a quarantined step stays in the flaky list")
		require.True(t, list.Steps[0].Quarantined)
		require.Equal(t, "waits on the shared database", list.Steps[0].Quarantine["note"])
		require.InDelta(t, 0.25, list.Steps[0].FlakeRate, 1e-9, "its figures are unchanged")
	})

	t.Run("a quarantined flaky step no longer makes its pipeline unhealthy", func(t *testing.T) {
		t.Parallel()

		// Every run passes; only the "test" step flakes inside them, so the
		// flaky-step signal is the pipeline's only reason to be unhealthy.
		srv := newTestServer(t, nil)
		repoID := seedRepo(t, srv)

		for i, stepConclusion := range []string{"success", "failure", "success", "success"} {
			ago := time.Duration(4-i) * time.Hour
			id := string(rune('a' + i))
			run := seedRecentRun(t, srv, repoID, "CI", id, "success", ago)
			seedJobStep(t, srv, run.ID, id+"-job", "test", stepConclusion)
		}

		_, body := getFlakySteps(t, srv, "")
		require.Len(t, body.Steps, 1)

		id := body.Steps[0].PipelineID

		health := func() string {
			rec := do(srv, http.MethodGet, "/api/pipelines/"+id, "")
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

			var out struct {
				HealthStatus string `json:"healthStatus"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))

			return out.HealthStatus
		}

		require.Equal(t, "unhealthy", health())
		require.Equal(t, http.StatusOK, do(srv, http.MethodPut, quarantinePath(id, "test"), `{}`).Code)
		require.Equal(t, "healthy", health())
	})

	t.Run("the pipeline's step list reports the mark", func(t *testing.T) {
		t.Parallel()

		srv, id, _ := flakyServer(t)
		require.Equal(t, http.StatusOK, do(srv, http.MethodPut, quarantinePath(id, "test"), `{"note":"n"}`).Code)

		rec := do(srv, http.MethodGet, "/api/pipelines/"+id+"/steps", "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var steps []struct {
			Name        string `json:"name"`
			Flaky       bool   `json:"flaky"`
			Quarantined bool   `json:"quarantined"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &steps))
		require.Len(t, steps, 1)
		require.True(t, steps[0].Flaky)
		require.True(t, steps[0].Quarantined)
	})

	t.Run("quarantining again renews it and replaces the note", func(t *testing.T) {
		t.Parallel()

		srv, id, _ := flakyServer(t)
		require.Equal(t, http.StatusOK, do(srv, http.MethodPut, quarantinePath(id, "test"), `{"note":"first"}`).Code)
		require.Equal(t, http.StatusOK, do(srv, http.MethodPut, quarantinePath(id, "test"), `{"note":"second"}`).Code)

		list := listQuarantined(t, srv)
		require.Equal(t, "second", list.Steps[0].Quarantine["note"])
	})

	t.Run("un-quarantining clears it", func(t *testing.T) {
		t.Parallel()

		srv, id, _ := flakyServer(t)
		require.Equal(t, http.StatusOK, do(srv, http.MethodPut, quarantinePath(id, "test"), `{}`).Code)

		rec := do(srv, http.MethodDelete, quarantinePath(id, "test"), "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var got quarantineBody
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.False(t, got.Quarantined)
		require.Nil(t, got.Quarantine)

		require.False(t, listQuarantined(t, srv).Steps[0].Quarantined)

		again := do(srv, http.MethodDelete, quarantinePath(id, "test"), "")
		require.Equal(t, http.StatusOK, again.Code, "clearing a step with no mark is fine")
	})

	t.Run("a note is at most 500 characters", func(t *testing.T) {
		t.Parallel()

		srv, id, _ := flakyServer(t)

		ok := do(srv, http.MethodPut, quarantinePath(id, "test"), `{"note":"`+strings.Repeat("é", 500)+`"}`)
		require.Equal(t, http.StatusOK, ok.Code, "500 characters is the limit, however many bytes")

		tooLong := do(srv, http.MethodPut, quarantinePath(id, "test"), `{"note":"`+strings.Repeat("a", 501)+`"}`)
		require.Equal(t, http.StatusUnprocessableEntity, tooLong.Code)
	})

	t.Run("an unknown or malformed pipeline is 404", func(t *testing.T) {
		t.Parallel()

		srv, _, _ := flakyServer(t)

		for _, method := range []string{http.MethodPut, http.MethodDelete} {
			require.Equal(t, http.StatusNotFound, do(srv, method, quarantinePath("not-a-pipeline-id", "test"), `{}`).Code, method)
		}
	})

	t.Run("a step name with a slash is one path segment", func(t *testing.T) {
		t.Parallel()

		srv, id, _ := flakyServer(t)
		require.Equal(t, http.StatusOK, do(srv, http.MethodPut, quarantinePath(id, "build/linux"), `{}`).Code)
	})

	t.Run("only a session may write; a token can read the mark", func(t *testing.T) {
		t.Parallel()

		srv, id, authStore := flakyServer(t)

		userID, err := authStore.Session(context.Background(), srv.sessionCookie.Value)
		require.NoError(t, err)

		_, raw, err := authStore.CreateToken(context.Background(), userID, 0)
		require.NoError(t, err)

		bearer := func(method, path, body string) *httptest.ResponseRecorder {
			req := httptest.NewRequest(method, path, strings.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+raw)

			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)

			return rec
		}

		require.Equal(t, http.StatusUnauthorized, bearer(http.MethodPut, quarantinePath(id, "test"), `{}`).Code)
		require.Equal(t, http.StatusUnauthorized, bearer(http.MethodDelete, quarantinePath(id, "test"), "").Code)
		require.False(t, listQuarantined(t, srv).Steps[0].Quarantined, "nothing was stored")

		require.Equal(t, http.StatusOK, do(srv, http.MethodPut, quarantinePath(id, "test"), `{"note":"seen"}`).Code)

		read := bearer(http.MethodGet, "/api/steps/flaky", "")
		require.Equal(t, http.StatusOK, read.Code)
		require.Contains(t, read.Body.String(), `"quarantined":true`)
		require.Contains(t, read.Body.String(), `"note":"seen"`)
	})
}
