package httpserver_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	apiBodyLimit     = 64 << 10
	webhookBodyLimit = 2 << 20
)

// jsonOfSize is a JSON object a little over n bytes: valid JSON, so a rejection
// has to come from the size and not from a parse error.
func jsonOfSize(n int) []byte {
	return []byte(`{"padding":"` + strings.Repeat("a", n) + `"}`)
}

func postBody(t *testing.T, srv testServer, method, path string, body []byte, authed bool) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if authed {
		req = srv.authenticated(req)
	}

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	return rec
}

func requirePayloadTooLarge(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code, rec.Body.String())
	require.JSONEq(t, `{"code":"payload_too_large","message":"request body too large"}`, rec.Body.String())
}

func TestBodyLimit_APIRoutesRejectOversizedBodies(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)

	cases := []struct {
		name, method, path string
		authed             bool
	}{
		{"register a repo", http.MethodPost, "/api/repos", true},
		{"discover repos", http.MethodPost, "/api/repos/discover", true},
		{"update settings", http.MethodPatch, "/api/settings", true},
		{"the MCP endpoint", http.MethodPost, "/api/mcp", true},
		{"the public login ceremony", http.MethodPost, "/api/auth/login", false},
		{"the public registration ceremony", http.MethodPost, "/api/auth/register", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			requirePayloadTooLarge(t, postBody(t, srv, tc.method, tc.path, jsonOfSize(apiBodyLimit), tc.authed))
		})
	}
}

func TestBodyLimit_ABodyWithNoDeclaredLengthIsStillCut(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)

	// A chunked request declares no Content-Length, so only reading it can
	// tell it's too big.
	req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", io.NopCloser(bytes.NewReader(jsonOfSize(apiBodyLimit)))))
	req.ContentLength = -1

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	requirePayloadTooLarge(t, rec)
}

func TestBodyLimit_AnOrdinaryBodyIsUnaffected(t *testing.T) {
	t.Parallel()

	srv := newSettingsTestServer(t)

	rec := postBody(t, srv, http.MethodPatch, "/api/settings", []byte(`{"theme":"dark"}`), true)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

func TestBodyLimit_WebhooksAllowMoreThanTheAPIButNotUnbounded(t *testing.T) {
	t.Parallel()

	srv := newTestServer(t, nil)

	post := func(body []byte) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/webhooks/github", bytes.NewReader(body))
		req.Header.Set("X-Hub-Signature-256", "sha256=00")
		req.Header.Set("X-GitHub-Event", "workflow_run")

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		return rec
	}

	// Over the API limit and under the webhook one: read and judged on its
	// signature (an untracked repo is a 401), not refused for its size.
	require.Equal(t, http.StatusUnauthorized, post(jsonOfSize(apiBodyLimit*4)).Code)

	requirePayloadTooLarge(t, post(jsonOfSize(webhookBodyLimit)))
}

func TestBodyLimit_FieldLengths(t *testing.T) {
	t.Parallel()

	srv := newSettingsTestServer(t)

	register := func(in map[string]string) int {
		body, err := json.Marshal(in)
		require.NoError(t, err)

		return postBody(t, srv, http.MethodPost, "/api/repos", body, true).Code
	}

	valid := func() map[string]string {
		return map[string]string{"forge": "github", "identifier": "alrayyes/pipeline-analytics", "token": "ghp_x"}
	}

	t.Run("an identifier past its limit is a 400", func(t *testing.T) {
		in := valid()
		in["identifier"] = strings.Repeat("a", 201)
		require.Equal(t, http.StatusBadRequest, register(in))
	})

	t.Run("a token past its limit is a 400", func(t *testing.T) {
		in := valid()
		in["token"] = strings.Repeat("t", 1025)
		require.Equal(t, http.StatusBadRequest, register(in))
	})

	t.Run("an instance URL past its limit is a 400", func(t *testing.T) {
		in := valid()
		in["forge"] = "forgejo"
		in["forgejoInstanceUrl"] = "https://" + strings.Repeat("a", 2048) + ".example"
		require.Equal(t, http.StatusBadRequest, register(in))
	})

	t.Run("a discovery token past its limit is a 400", func(t *testing.T) {
		body, err := json.Marshal(map[string]string{"forge": "github", "token": strings.Repeat("t", 1025)})
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, postBody(t, srv, http.MethodPost, "/api/repos/discover", body, true).Code)
	})

	t.Run("a repo selector past its limit is a 400", func(t *testing.T) {
		body, err := json.Marshal(map[string]string{"pipelinesRepoSelector": strings.Repeat("r", 129)})
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, postBody(t, srv, http.MethodPatch, "/api/settings", body, true).Code)
	})

	t.Run("values at the limit are accepted", func(t *testing.T) {
		body, err := json.Marshal(map[string]string{"pipelinesRepoSelector": strings.Repeat("r", 128)})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, postBody(t, srv, http.MethodPatch, "/api/settings", body, true).Code)
	})
}
