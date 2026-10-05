package github

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

// maxLogBytes bounds how much of one log is read while looking for its tail.
// Past it the lines kept are the last ones seen before the cap, and the
// result is marked truncated.
const maxLogBytes = 64 << 20

var _ ingestion.JobLogReader = (*Client)(nil)

// JobLogTail implements ingestion.JobLogReader. GitHub answers
// GET .../actions/jobs/{id}/logs with a 302 to a plain-text download that
// expires after a minute
// (https://docs.github.com/en/rest/actions/workflow-jobs), so the redirect
// is followed here and the download URL never leaves the server. net/http
// drops the Authorization header on a cross-host redirect, which is what the
// download wants.
func (c *Client) JobLogTail(ctx context.Context, req ingestion.JobLogRequest) (ingestion.JobLogTail, error) {
	owner, name, err := splitIdentifier(req.Identifier)
	if err != nil {
		return ingestion.JobLogTail{}, err
	}

	base := defaultBaseURL
	if c.baseURL != nil {
		base = c.baseURL.String()
	}

	url := fmt.Sprintf("%s/repos/%s/%s/actions/jobs/%s/logs", strings.TrimRight(base, "/"), owner, name, req.ForgeJobID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ingestion.JobLogTail{}, fmt.Errorf("build request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+req.Token)
	httpReq.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ingestion.JobLogTail{}, fmt.Errorf("%w: %w", ingestion.ErrLogUnreachable, err)
	}
	defer func() { _ = resp.Body.Close() }()

	c.recordRateLimit(req.Token, resp.Header)

	switch resp.StatusCode {
	case http.StatusOK:
		return tailLines(resp.Body, req.Lines)
	case http.StatusNotFound, http.StatusGone:
		return ingestion.JobLogTail{}, ingestion.ErrLogExpired
	case http.StatusUnauthorized, http.StatusForbidden:
		return ingestion.JobLogTail{}, ingestion.ErrLogForbidden
	default:
		return ingestion.JobLogTail{}, fmt.Errorf("%w: status %d", ingestion.ErrLogUnreachable, resp.StatusCode)
	}
}

// tailLines streams r and keeps the last n lines.
func tailLines(r io.Reader, n int) (ingestion.JobLogTail, error) {
	if n < 1 {
		n = 1
	}

	reader := bufio.NewReaderSize(io.LimitReader(r, maxLogBytes), 64<<10)
	ring := make([]string, 0, n)
	total := 0

	for {
		line, err := readLine(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return ingestion.JobLogTail{}, fmt.Errorf("%w: read log: %w", ingestion.ErrLogUnreachable, err)
		}

		total++

		if len(ring) == n {
			copy(ring, ring[1:])
			ring = ring[:n-1]
		}

		ring = append(ring, line)
	}

	return ingestion.JobLogTail{Lines: ring, Truncated: total > n}, nil
}

// readLine reads one line without its newline, cut to MaxLogLineLength and
// to a rune boundary, and discards the rest of an overlong one.
func readLine(r *bufio.Reader) (string, error) {
	var line []byte

	for {
		part, isPrefix, err := r.ReadLine()
		if err != nil {
			return "", err //nolint:wrapcheck // io.EOF is the caller's loop signal
		}

		if room := ingestion.MaxLogLineLength - len(line); room > 0 {
			line = append(line, part[:min(room, len(part))]...)
		}

		if !isPrefix {
			return finishLine(line), nil
		}
	}
}

func finishLine(line []byte) string {
	line = bytes.TrimSuffix(line, []byte("\r"))

	for len(line) > 0 && !utf8.Valid(line) && len(line) == ingestion.MaxLogLineLength {
		line = line[:len(line)-1]
	}

	return string(line)
}
