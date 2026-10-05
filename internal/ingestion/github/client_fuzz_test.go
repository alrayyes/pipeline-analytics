package github_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	ghclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/github"
)

// GitHub's responses are external data. The client may return an error for
// any body, never panic.
func FuzzListRecentRunsResponses(f *testing.F) {
	f.Add(workflowRunsPayload, workflowJobsPayload)
	f.Add(workflowRunsPayload, `{"jobs": [{"steps": [null]}]}`)
	f.Add(`{"workflow_runs": [null]}`, `{}`)
	f.Add(`not json`, `not json`)

	f.Fuzz(func(t *testing.T, runs, jobs string) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/repos/alrayyes/pipeline-analytics/actions/runs" {
				_, _ = w.Write([]byte(runs))

				return
			}

			_, _ = w.Write([]byte(jobs))
		}))
		defer server.Close()

		client, err := ghclient.NewClient(server.URL + "/")
		if err != nil {
			t.Fatal(err)
		}

		_, _ = client.ListRecentRuns(context.Background(), ingestion.ListRunsRequest{
			Identifier: "alrayyes/pipeline-analytics",
			Token:      "ghp_test",
		})
	})
}
