package forgejo_test

import (
	"context"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	forgejoclient "github.com/alrayyes/pipeline-analytics/internal/ingestion/forgejo"
	"github.com/stretchr/testify/require"
)

// Forgejo's Actions log API landed in v16; the instance this service targets
// runs 11.0.16, so there is nothing to call and the client says so without
// making a request.
func TestClient_JobLogTail_Unsupported(t *testing.T) {
	t.Parallel()

	client, err := forgejoclient.NewClient()
	require.NoError(t, err)

	_, err = client.JobLogTail(context.Background(), ingestion.JobLogRequest{
		InstanceURL: "https://forgejo.example",
		Identifier:  "alrayyes/dotfiles",
		Token:       "t",
		ForgeJobID:  "7",
		Lines:       200,
	})
	require.ErrorIs(t, err, ingestion.ErrLogUnsupported)
}
