package forgejo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	forgejoclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/forgejo"
	"github.com/stretchr/testify/require"
)

// A null where an object belongs is valid JSON that decodes to a nil pointer,
// so a reply from a broken or hostile instance must not be dereferenced
// unchecked (#446).
func TestClient_ListRecentRuns_SkipsNullEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		runs, jobs string
		wantRuns   int
		wantJobs   int
		wantSteps  int
	}{
		{"a null run", `{"workflow_runs": [null]}`, `[]`, 0, 0, 0},
		{"a null job in a bare array", actionRunsPayload, `[null]`, -1, 0, 0},
		{"a null job in a wrapped response", actionRunsPayload, `{"jobs": [null]}`, -1, 0, 0},
		{"a null step", actionRunsPayload, `[{"id": 1, "name": "build", "steps": [null]}]`, -1, 1, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if r.URL.Path == "/api/v1/repos/alrayyes/dotfiles/actions/runs" {
					_, _ = w.Write([]byte(tc.runs))

					return
				}

				_, _ = w.Write([]byte(tc.jobs))
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

			if tc.wantRuns >= 0 {
				require.Len(t, result.Runs, tc.wantRuns)
			}

			if tc.wantRuns == 0 {
				return
			}

			require.NotEmpty(t, result.Runs)
			require.Len(t, result.Runs[0].Jobs, tc.wantJobs)

			if tc.wantJobs > 0 {
				require.Len(t, result.Runs[0].Jobs[0].Steps, tc.wantSteps)
			}
		})
	}
}
