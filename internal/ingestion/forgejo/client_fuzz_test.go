package forgejo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	forgejoclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/forgejo"
)

// A Forgejo instance is someone else's server, so what it returns is
// external data. The client may return an error for any body, never panic.
func FuzzListRecentRunsResponses(f *testing.F) {
	f.Add(actionRunsPayload, actionJobsPayloadWrapped)
	f.Add(actionRunsPayload, actionJobsPayloadBareArray)
	f.Add(actionRunsPayload, `{"jobs": null}`)
	f.Add(actionRunsPayload, `[{"steps": [null]}]`)
	f.Add(`{"workflow_runs": [null]}`, `[]`)
	f.Add(`not json`, `not json`)

	f.Fuzz(func(t *testing.T, runs, jobs string) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			if r.URL.Path == "/api/v1/repos/alrayyes/dotfiles/actions/runs" {
				_, _ = w.Write([]byte(runs))

				return
			}

			_, _ = w.Write([]byte(jobs))
		}))
		defer server.Close()

		client, err := forgejoclient.NewClient()
		if err != nil {
			t.Fatal(err)
		}

		_, _ = client.ListRecentRuns(context.Background(), ingestion.ListRunsRequest{
			InstanceURL: server.URL,
			Identifier:  "alrayyes/dotfiles",
			Token:       "forgejo_test_token",
		})
	})
}
