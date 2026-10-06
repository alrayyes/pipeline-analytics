package github_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ghclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/github"
	"github.com/stretchr/testify/require"
)

func actionRequest(failedOnly bool) ingestion.RunActionRequest {
	return ingestion.RunActionRequest{
		Identifier: "alrayyes/pipeline-analytics",
		Token:      "ghp_test",
		ForgeRunID: "77",
		FailedOnly: failedOnly,
	}
}

// answering returns a client whose forge answers every request with status,
// and a pointer to the "METHOD path" and auth header of the last one.
func answering(t *testing.T, status int) (*ghclient.Client, *[2]string) {
	t.Helper()

	var seen [2]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = [2]string{r.Method + " " + r.URL.Path, r.Header.Get("Authorization")}

		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)

	client, err := ghclient.NewClient(server.URL + "/")
	require.NoError(t, err)

	return client, &seen
}

func TestClient_RerunRun(t *testing.T) {
	t.Parallel()

	t.Run("re-runs only the failed jobs when asked to", func(t *testing.T) {
		t.Parallel()

		client, seen := answering(t, http.StatusCreated)

		require.NoError(t, client.RerunRun(context.Background(), actionRequest(true)))
		require.Equal(t, [2]string{"POST /repos/alrayyes/pipeline-analytics/actions/runs/77/rerun-failed-jobs", "Bearer ghp_test"}, *seen)
	})

	t.Run("re-runs the whole run otherwise", func(t *testing.T) {
		t.Parallel()

		client, seen := answering(t, http.StatusCreated)

		require.NoError(t, client.RerunRun(context.Background(), actionRequest(false)))
		require.Equal(t, "POST /repos/alrayyes/pipeline-analytics/actions/runs/77/rerun", seen[0])
	})

	t.Run("a read-only token is forbidden", func(t *testing.T) {
		t.Parallel()

		client, _ := answering(t, http.StatusForbidden)

		require.ErrorIs(t, client.RerunRun(context.Background(), actionRequest(false)), ingestion.ErrActionForbidden)
	})

	t.Run("a run the forge won't re-run is not actionable", func(t *testing.T) {
		t.Parallel()

		client, _ := answering(t, http.StatusConflict)

		require.ErrorIs(t, client.RerunRun(context.Background(), actionRequest(false)), ingestion.ErrNotActionable)
	})

	t.Run("a forge error is unreachable", func(t *testing.T) {
		t.Parallel()

		client, _ := answering(t, http.StatusBadGateway)

		require.ErrorIs(t, client.RerunRun(context.Background(), actionRequest(false)), ingestion.ErrActionUnreachable)
	})
}

func TestClient_CancelRun(t *testing.T) {
	t.Parallel()

	t.Run("posts to the cancel endpoint", func(t *testing.T) {
		t.Parallel()

		client, seen := answering(t, http.StatusAccepted)

		require.NoError(t, client.CancelRun(context.Background(), actionRequest(false)))
		require.Equal(t, "POST /repos/alrayyes/pipeline-analytics/actions/runs/77/cancel", seen[0])
	})

	t.Run("a read-only token is forbidden", func(t *testing.T) {
		t.Parallel()

		client, _ := answering(t, http.StatusForbidden)

		require.ErrorIs(t, client.CancelRun(context.Background(), actionRequest(false)), ingestion.ErrActionForbidden)
	})
}
