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

const actionRunsPayload = `{
  "total_count": 1,
  "workflow_runs": [
    {
      "id": 42,
      "path": ".gitea/workflows/ci.yml",
      "status": "completed",
      "conclusion": "success",
      "started_at": "2026-09-15T10:00:00Z",
      "completed_at": "2026-09-15T10:05:00Z",
      "html_url": "https://forgejo.example.com/alrayyes/dotfiles/actions/runs/42"
    }
  ]
}`

const actionJobsPayload = `{
  "total_count": 1,
  "jobs": [
    {
      "id": 100,
      "run_id": 42,
      "name": "build",
      "status": "completed",
      "conclusion": "success",
      "created_at": "2026-09-15T09:59:00Z",
      "started_at": "2026-09-15T10:00:00Z",
      "completed_at": "2026-09-15T10:04:00Z",
      "html_url": "https://forgejo.example.com/alrayyes/dotfiles/actions/runs/42/jobs/100",
      "steps": [
        {"name": "checkout", "status": "completed", "conclusion": "success", "number": 1}
      ]
    }
  ]
}`

func TestClient_ListRecentRuns(t *testing.T) {
	t.Parallel()

	t.Run("fetches recent runs with their jobs and steps", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Path {
			case "/api/v1/repos/alrayyes/dotfiles/actions/runs":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(actionRunsPayload))
			case "/api/v1/repos/alrayyes/dotfiles/actions/runs/42/jobs":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(actionJobsPayload))
			default:
				t.Errorf("unexpected request path: %s", r.URL.Path)
			}
		}))
		defer server.Close()

		client, err := forgejoclient.NewClient()
		require.NoError(t, err)

		result, err := client.ListRecentRuns(context.Background(), ingestion.ListRunsRequest{
			InstanceURL: server.URL,
			Identifier:  "alrayyes/dotfiles",
			Token:       "forgejo_test_token",
		})
		require.NoError(t, err)

		require.False(t, result.NotModified)
		require.Empty(t, result.ETag) // Forgejo has no conditional-request support here
		require.Len(t, result.Runs, 1)

		run := result.Runs[0]
		require.Equal(t, "42", run.ForgeRunID)
		require.Equal(t, "ci", run.PipelineName)
		require.Equal(t, "success", run.Conclusion)
		require.Len(t, run.Jobs, 1)

		job := run.Jobs[0]
		require.Equal(t, "100", job.ForgeJobID)
		require.Equal(t, "build", job.Name)
		require.Len(t, job.Steps, 1)
		require.Equal(t, "checkout", job.Steps[0].Name)
	})

	t.Run("rejects an identifier not in owner/name form", func(t *testing.T) {
		t.Parallel()

		client, err := forgejoclient.NewClient()
		require.NoError(t, err)

		_, err = client.ListRecentRuns(context.Background(), ingestion.ListRunsRequest{InstanceURL: "https://example.com", Identifier: "not-owner-slash-name"})
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

		_, err = client.ListRecentRuns(context.Background(), ingestion.ListRunsRequest{InstanceURL: server.URL, Identifier: "alrayyes/dotfiles", Token: "t"})
		require.Error(t, err)
	})
}

func TestClient_ListAccessibleRepos(t *testing.T) {
	t.Parallel()

	t.Run("sends an authenticated request and returns full_name identifiers", func(t *testing.T) {
		t.Parallel()

		var gotPath, gotAuth string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotAuth = r.Header.Get("Authorization")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"full_name": "alrayyes/dotfiles"}, {"full_name": "alrayyes/pipeline-analytics"}]`))
		}))
		defer server.Close()

		client, err := forgejoclient.NewClient()
		require.NoError(t, err)

		repos, err := client.ListAccessibleRepos(context.Background(), ingestion.ListAccessibleReposRequest{
			InstanceURL: server.URL,
			Token:       "forgejo_test_token",
		})
		require.NoError(t, err)

		require.Equal(t, "/api/v1/user/repos", gotPath)
		require.Equal(t, "token forgejo_test_token", gotAuth)
		require.Equal(t, []string{"alrayyes/dotfiles", "alrayyes/pipeline-analytics"}, repos)
	})

	t.Run("wraps a forge-side failure", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client, err := forgejoclient.NewClient()
		require.NoError(t, err)

		_, err = client.ListAccessibleRepos(context.Background(), ingestion.ListAccessibleReposRequest{
			InstanceURL: server.URL,
			Token:       "bad",
		})
		require.Error(t, err)
	})
}
