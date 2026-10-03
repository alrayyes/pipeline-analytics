package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Request body limits. A handler that decodes whatever it is sent lets one
// oversized request exhaust memory (OWASP API4), so every body is capped
// before a handler reads it. The API and the MCP endpoint only ever carry
// small JSON; the forge webhook receivers get more room, since a delivery's
// size is the forge's to choose, but not unbounded.
const (
	maxAPIBodyBytes     = 64 << 10
	maxWebhookBodyBytes = 2 << 20
)

// Field limits for the free-text request fields, matching the maxLength in
// openapi/openapi.yaml's request schemas.
const (
	maxIdentifierLength  = 200
	maxTokenLength       = 1024
	maxInstanceURLLength = 2048
)

func bodyLimitFor(r *http.Request) int64 {
	if strings.HasPrefix(r.URL.Path, "/webhooks/") {
		return maxWebhookBodyBytes
	}

	return maxAPIBodyBytes
}

func writePayloadTooLarge(w http.ResponseWriter) {
	writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "request body too large")
}

// isBodyTooLarge reports whether err came from a body read hitting its cap.
func isBodyTooLarge(err error) bool {
	var tooLarge *http.MaxBytesError

	return errors.As(err, &tooLarge)
}

// limitBodies caps every request body. A declared Content-Length over the
// limit is refused without reading anything; a body that declares none (or
// lies) is cut by MaxBytesReader at the limit, and the handler that reads it
// turns that into the same 413.
func limitBodies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := bodyLimitFor(r)

		if r.ContentLength > limit {
			writePayloadTooLarge(w)

			return
		}

		if r.Body != nil && r.Body != http.NoBody {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
		}

		next.ServeHTTP(w, r)
	})
}

// readJSON decodes the request's JSON body into dst. On failure it has already
// written the response (413 for a body past its cap, 400 for anything
// malformed) and returns false.
func readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		if isBodyTooLarge(err) {
			writePayloadTooLarge(w)
		} else {
			writeError(w, http.StatusBadRequest, "invalid_body", "malformed JSON body")
		}

		return false
	}

	return true
}
