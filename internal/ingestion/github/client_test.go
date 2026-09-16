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

const workflowRunsPayload = `{
  "workflow_runs": [
    {
      "id": 1001,
      "name": "CI",
      "status": "completed",
      "conclusion": "success",
      "run_started_at": "2026-09-15T10:00:00Z",
      "updated_at": "2026-09-15T10:05:00Z",
      "html_url": "https://github.com/alrayyes/pipeline-analytics/actions/runs/1001"
    }
  ]
}`

const workflowJobsPayload = `{
  "jobs": [
    {
      "id": 5001,
      "name": "build",
      "status": "completed",
      "conclusion": "success",
      "created_at": "2026-09-15T09:59:00Z",
      "started_at": "2026-09-15T10:00:00Z",
      "completed_at": "2026-09-15T10:04:00Z",
      "html_url": "https://github.com/alrayyes/pipeline-analytics/actions/runs/1001/job/5001",
      "steps": [
        {"name": "checkout", "status": "completed", "conclusion": "success", "number": 1}
      ]
    }
  ]
}`

func TestClient_ListRecentRuns(t *testing.T) {
	t.Parallel()

	t.Run("a first poll (no ETag) fetches runs with their jobs and steps", func(t *testing.T) {
		t.Parallel()

		var gotIfNoneMatch string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/repos/alrayyes/pipeline-analytics/actions/runs":
				gotIfNoneMatch = r.Header.Get("If-None-Match")
				w.Header().Set("ETag", `"v2"`)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(workflowRunsPayload))
			case "/repos/alrayyes/pipeline-analytics/actions/runs/1001/jobs":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(workflowJobsPayload))
			default:
				t.Errorf("unexpected request path: %s", r.URL.Path)
			}
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		result, err := client.ListRecentRuns(context.Background(), ingestion.ListRunsRequest{
			Identifier: "alrayyes/pipeline-analytics",
			Token:      "ghp_test",
		})
		require.NoError(t, err)

		require.Empty(t, gotIfNoneMatch)
		require.False(t, result.NotModified)
		require.Equal(t, `"v2"`, result.ETag)
		require.Len(t, result.Runs, 1)

		run := result.Runs[0]
		require.Equal(t, "1001", run.ForgeRunID)
		require.Equal(t, "CI", run.PipelineName)
		require.Equal(t, "success", run.Conclusion)
		require.Len(t, run.Jobs, 1)

		job := run.Jobs[0]
		require.Equal(t, "5001", job.ForgeJobID)
		require.Equal(t, "build", job.Name)
		require.Len(t, job.Steps, 1)
		require.Equal(t, "checkout", job.Steps[0].Name)
	})

	t.Run("sends the repo's ETag as If-None-Match and reports 304 as unmodified", func(t *testing.T) {
		t.Parallel()

		var gotIfNoneMatch string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/repos/alrayyes/pipeline-analytics/actions/runs/1001/jobs" {
				t.Fatal("jobs should never be fetched for an unmodified run list")
			}

			gotIfNoneMatch = r.Header.Get("If-None-Match")
			w.WriteHeader(http.StatusNotModified)
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		result, err := client.ListRecentRuns(context.Background(), ingestion.ListRunsRequest{
			Identifier: "alrayyes/pipeline-analytics",
			Token:      "ghp_test",
			ETag:       `"v1"`,
		})
		require.NoError(t, err)

		require.Equal(t, `"v1"`, gotIfNoneMatch)
		require.True(t, result.NotModified)
		require.Empty(t, result.Runs)
	})

	t.Run("rejects an identifier not in owner/name form", func(t *testing.T) {
		t.Parallel()

		client, err := ghclient.NewClient("")
		require.NoError(t, err)

		_, err = client.ListRecentRuns(context.Background(), ingestion.ListRunsRequest{Identifier: "not-owner-slash-name"})
		require.ErrorIs(t, err, ghclient.ErrInvalidIdentifier)
	})

	t.Run("wraps an unexpected status from the forge", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		_, err = client.ListRecentRuns(context.Background(), ingestion.ListRunsRequest{Identifier: "alrayyes/pipeline-analytics", Token: "ghp_test"})
		require.Error(t, err)
	})
}

func TestClient_ListAccessibleRepos(t *testing.T) {
	t.Parallel()

	t.Run("sends an authenticated request and returns full_name identifiers", func(t *testing.T) {
		t.Parallel()

		var gotPath, gotAuth, gotQuery string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotAuth = r.Header.Get("Authorization")
			gotQuery = r.URL.RawQuery

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"full_name": "alrayyes/pipeline-analytics"}, {"full_name": "alrayyes/dotfiles"}]`))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		repos, err := client.ListAccessibleRepos(context.Background(), ingestion.ListAccessibleReposRequest{
			Token: "ghp_test",
		})
		require.NoError(t, err)

		require.Equal(t, "/user/repos", gotPath)
		require.Equal(t, "Bearer ghp_test", gotAuth)
		require.Contains(t, gotQuery, "sort=pushed")
		require.Equal(t, []string{"alrayyes/pipeline-analytics", "alrayyes/dotfiles"}, repos)
	})

	t.Run("wraps a forge-side failure", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		_, err = client.ListAccessibleRepos(context.Background(), ingestion.ListAccessibleReposRequest{Token: "bad"})
		require.Error(t, err)
	})

	t.Run("follows pagination across multiple pages", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if r.URL.Query().Get("page") == "2" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[{"full_name": "alrayyes/page-two-repo"}]`))

				return
			}

			// Rewrite (not append to) the query so this doesn't produce a
			// second, shadowed "page" param if the request already carried
			// one. Headers also have to be set before WriteHeader -- once
			// that's called, the response is flushed and a later
			// Header().Set is silently ignored.
			nextURL := *r.URL
			q := nextURL.Query()
			q.Set("page", "2")
			nextURL.RawQuery = q.Encode()
			w.Header().Set("Link", `<`+nextURL.String()+`>; rel="next"`)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"full_name": "alrayyes/page-one-repo"}]`))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		repos, err := client.ListAccessibleRepos(context.Background(), ingestion.ListAccessibleReposRequest{Token: "ghp_test"})
		require.NoError(t, err)

		require.Equal(t, []string{"alrayyes/page-one-repo", "alrayyes/page-two-repo"}, repos)
	})

	t.Run("excludes archived and forked repos", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"full_name": "alrayyes/normal-repo", "archived": false, "fork": false},
				{"full_name": "alrayyes/archived-repo", "archived": true, "fork": false},
				{"full_name": "alrayyes/forked-repo", "archived": false, "fork": true}
			]`))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		require.NoError(t, err)

		repos, err := client.ListAccessibleRepos(context.Background(), ingestion.ListAccessibleReposRequest{Token: "ghp_test"})
		require.NoError(t, err)

		require.Equal(t, []string{"alrayyes/normal-repo"}, repos)
	})
}
