package forgejo

import (
	"context"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

var _ ingestion.JobLogReader = (*Client)(nil)

// JobLogTail implements ingestion.JobLogReader. Forgejo's Actions log REST
// API landed in v16 (forgejo/forgejo#12666); earlier versions serve logs only
// through the web UI, behind a session. The instance this service targets
// runs 11.0.16, so there is nothing to call: report it, and let the caller
// show the deep link. Revisit when a v16 instance is a supported target.
func (c *Client) JobLogTail(context.Context, ingestion.JobLogRequest) (ingestion.JobLogTail, error) {
	return ingestion.JobLogTail{}, ingestion.ErrLogUnsupported
}
