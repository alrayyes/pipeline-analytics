package github_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ghclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateWebhook(t *testing.T) {
	t.Parallel()

	t.Run("sends an authenticated request creating a workflow webhook", func(t *testing.T) {
		t.Parallel()

		var gotMethod, gotPath, gotAuth, gotBody string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotAuth = r.Header.Get("Authorization")

			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			gotBody = string(body)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id": 42}`))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		err = client.CreateWebhook(context.Background(), ingestion.CreateWebhookRequest{
			Identifier:  "alrayyes/pipeline-analytics",
			Token:       "ghp_test",
			CallbackURL: "https://example.com/webhooks/github",
			Secret:      "shh",
		})
		require.NoError(t, err)

		require.Equal(t, http.MethodPost, gotMethod)
		require.Equal(t, "/repos/alrayyes/pipeline-analytics/hooks", gotPath)
		require.Equal(t, "Bearer ghp_test", gotAuth)
		require.Contains(t, gotBody, "workflow_job")
		require.Contains(t, gotBody, "https://example.com/webhooks/github")
	})

	t.Run("rejects an identifier not in owner/name form", func(t *testing.T) {
		t.Parallel()

		client, err := ghclient.NewClient("")
		require.NoError(t, err)

		err = client.CreateWebhook(context.Background(), ingestion.CreateWebhookRequest{
			Identifier: "not-owner-slash-name",
			Token:      "ghp_test",
		})
		require.ErrorIs(t, err, ghclient.ErrInvalidIdentifier)
	})

	t.Run("wraps a forge-side failure", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message": "insufficient scope"}`))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		err = client.CreateWebhook(context.Background(), ingestion.CreateWebhookRequest{
			Identifier: "alrayyes/pipeline-analytics",
			Token:      "ghp_test",
		})
		require.Error(t, err)
	})
}
