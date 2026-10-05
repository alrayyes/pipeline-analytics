package ingestion

import (
	"context"
	"errors"
)

// Why a forge couldn't supply a job's log. Each maps to a JobLog `reason` in
// the API (openspec/changes/api-job-log).
var (
	// ErrLogUnsupported means the forge has no log API: Forgejo before v16.
	ErrLogUnsupported = errors.New("forge has no job log API")
	// ErrLogExpired means the forge no longer has this log.
	ErrLogExpired = errors.New("job log expired")
	// ErrLogForbidden means the stored token can't read the log.
	ErrLogForbidden = errors.New("token can't read the job log")
	// ErrLogUnreachable means the forge didn't answer.
	ErrLogUnreachable = errors.New("forge unreachable")
)

// JobLogRequest names one job's log on its forge.
type JobLogRequest struct {
	// InstanceURL is set for Forgejo, empty for GitHub.
	InstanceURL string
	Identifier  string
	Token       string
	ForgeJobID  string
	// Lines is how many trailing lines to return.
	Lines int
}

// JobLogTail is the end of a job's log.
type JobLogTail struct {
	// Lines are oldest first, raw (ANSI included), each cut to MaxLogLineLength.
	Lines []string
	// Truncated is true when the log had more lines than were returned.
	Truncated bool
}

// MaxLogLineLength bounds one returned line, in bytes.
const MaxLogLineLength = 4096

// JobLogReader fetches a job's log from its forge on demand. It is separate
// from ForgeClient so a forge, or a test fake, that can't read logs doesn't
// have to say so. It returns one of the ErrLog* errors, wrapped, when it
// can't supply the log.
type JobLogReader interface {
	JobLogTail(ctx context.Context, req JobLogRequest) (JobLogTail, error)
}
