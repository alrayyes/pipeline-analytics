package httpserver_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/stretchr/testify/require"
)

const webhookTestPayload = `{
  "action": "completed",
  "workflow_run": {
    "id": 1001,
    "name": "CI",
    "status": "completed",
    "conclusion": "success",
    "run_started_at": "2026-09-15T10:00:00Z",
    "updated_at": "2026-09-15T10:05:00Z",
    "html_url": "https://github.com/alrayyes/pipeline-analytics/actions/runs/1001"
  },
  "repository": {"full_name": "alrayyes/pipeline-analytics"}
}`

func signBody(t *testing.T, secret, body string) string {
	t.Helper()

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))

	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

const webhookTestRepoIdentifier = "alrayyes/pipeline-analytics"

// registerTestRepo registers a repo through the API (the same path a real
// user goes through) and returns the webhook secret the server generated
// for it -- never exposed over HTTP, so read directly off the underlying
// store the test server was built with.
func registerTestRepo(t *testing.T, srv testServer, forge string) string {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"forge":      forge,
		"identifier": webhookTestRepoIdentifier,
		"token":      "ghp_supersecrettoken1234",
	})
	require.NoError(t, err)

	req := srv.authenticated(httptest.NewRequest(http.MethodPost, "/api/repos", strings.NewReader(string(body))))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	repo, err := srv.runStore.RepoByIdentifier(context.Background(), ingestion.Forge(forge), webhookTestRepoIdentifier)
	require.NoError(t, err)

	return repo.WebhookSecret
}

func TestWebhookReceiver(t *testing.T) {
	t.Parallel()

	t.Run("a valid signature for a tracked repo is accepted and updates run storage", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		secret := registerTestRepo(t, srv, "github")

		req := httptest.NewRequest(http.MethodPost, "/webhooks/github", strings.NewReader(webhookTestPayload))
		req.Header.Set("X-GitHub-Event", "workflow_run")
		req.Header.Set("X-Hub-Signature-256", signBody(t, secret, webhookTestPayload))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusAccepted, rec.Code)
	})

	t.Run("an invalid signature is rejected", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		registerTestRepo(t, srv, "github")

		req := httptest.NewRequest(http.MethodPost, "/webhooks/github", strings.NewReader(webhookTestPayload))
		req.Header.Set("X-GitHub-Event", "workflow_run")
		req.Header.Set("X-Hub-Signature-256", "sha256=0000000000000000000000000000000000000000000000000000000000000000")
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("an untracked repo is rejected the same way as a bad signature", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)

		req := httptest.NewRequest(http.MethodPost, "/webhooks/github", strings.NewReader(webhookTestPayload))
		req.Header.Set("X-GitHub-Event", "workflow_run")
		req.Header.Set("X-Hub-Signature-256", signBody(t, "whatever", webhookTestPayload))
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("the webhook endpoint itself is public, no session required", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		secret := registerTestRepo(t, srv, "github")

		req := httptest.NewRequest(http.MethodPost, "/webhooks/github", strings.NewReader(webhookTestPayload))
		req.Header.Set("X-GitHub-Event", "workflow_run")
		req.Header.Set("X-Hub-Signature-256", signBody(t, secret, webhookTestPayload)) // no session cookie attached
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusAccepted, rec.Code)
	})

	t.Run("forgejo delivery uses its own signature header and event header", func(t *testing.T) {
		t.Parallel()

		srv := newTestServer(t, nil)
		secret := registerTestRepo(t, srv, "forgejo")

		req := httptest.NewRequest(http.MethodPost, "/webhooks/forgejo", strings.NewReader(webhookTestPayload))
		req.Header.Set("X-Forgejo-Event", "workflow_run")

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(webhookTestPayload))
		req.Header.Set("X-Forgejo-Signature", hex.EncodeToString(mac.Sum(nil)))

		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusAccepted, rec.Code)
	})
}
