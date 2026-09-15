package forgejo_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	forgejoclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/forgejo"
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
			_, _ = w.Write([]byte(`{"id": 7}`))
		}))
		defer server.Close()

		client, err := forgejoclient.NewClient()
		require.NoError(t, err)

		err = client.CreateWebhook(context.Background(), ingestion.CreateWebhookRequest{
			InstanceURL: server.URL,
			Identifier:  "alrayyes/dotfiles",
			Token:       "forgejo_test_token",
			CallbackURL: "https://example.com/webhooks/forgejo",
			Secret:      "shh",
		})
		require.NoError(t, err)

		require.Equal(t, http.MethodPost, gotMethod)
		require.Equal(t, "/api/v1/repos/alrayyes/dotfiles/hooks", gotPath)
		require.Equal(t, "token forgejo_test_token", gotAuth)
		require.Contains(t, gotBody, "workflow_job")
		require.Contains(t, gotBody, "https://example.com/webhooks/forgejo")
	})

	t.Run("rejects an identifier not in owner/name form", func(t *testing.T) {
		t.Parallel()

		client, err := forgejoclient.NewClient()
		require.NoError(t, err)

		err = client.CreateWebhook(context.Background(), ingestion.CreateWebhookRequest{
			InstanceURL: "https://example.com",
			Identifier:  "not-owner-slash-name",
			Token:       "t",
		})
		require.ErrorIs(t, err, forgejoclient.ErrInvalidIdentifier)
	})

	t.Run("wraps a forge-side failure", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message": "insufficient scope"}`))
		}))
		defer server.Close()

		client, err := forgejoclient.NewClient()
		require.NoError(t, err)

		err = client.CreateWebhook(context.Background(), ingestion.CreateWebhookRequest{
			InstanceURL: server.URL,
			Identifier:  "alrayyes/dotfiles",
			Token:       "t",
		})
		require.Error(t, err)
	})
}
