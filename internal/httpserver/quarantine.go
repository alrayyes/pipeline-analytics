package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
)

// quarantineDTO is a quarantine mark as the API reports it. It is absent from
// a step that isn't quarantined.
type quarantineDTO struct {
	Note          string    `json:"note"`
	QuarantinedAt time.Time `json:"quarantinedAt"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

func toQuarantineDTO(q *metrics.Quarantine) *quarantineDTO {
	if q == nil {
		return nil
	}

	return &quarantineDTO{Note: q.Note, QuarantinedAt: q.QuarantinedAt, ExpiresAt: q.ExpiresAt}
}

type quarantineRequest struct {
	Note string `json:"note"`
}

// stepQuarantineDTO answers the quarantine writes: whether the step is now
// quarantined and, if so, the mark.
type stepQuarantineDTO struct {
	Quarantined bool           `json:"quarantined"`
	Quarantine  *quarantineDTO `json:"quarantine,omitempty"`
}

// quarantineHandler marks a flaky step as known. Both routes are session-only,
// like API tokens and saved forge tokens: a token or an agent can read a mark
// (the flaky and step lists carry it) but never set or clear one. Neither
// touches the forge.
type quarantineHandler struct {
	service *metrics.Service
}

// mountQuarantine registers the two session-only quarantine routes.
func mountQuarantine(mux *http.ServeMux, service *metrics.Service) {
	h := &quarantineHandler{service: service}
	mux.HandleFunc("PUT /api/pipelines/{pipelineId}/steps/{step}/quarantine", h.quarantine)
	mux.HandleFunc("DELETE /api/pipelines/{pipelineId}/steps/{step}/quarantine", h.unquarantine)
}

// quarantine answers PUT /api/pipelines/{pipelineId}/steps/{step}/quarantine.
// Marking a step that is already quarantined renews it.
func (h *quarantineHandler) quarantine(w http.ResponseWriter, r *http.Request) {
	if !requireViaSession(w, r, "quarantine a step") {
		return
	}

	var in quarantineRequest
	if !readOptionalJSON(w, r, &in) {
		return
	}

	q, err := h.service.QuarantineStep(r.Context(), metrics.PipelineID(r.PathValue("pipelineId")), r.PathValue("step"), in.Note, time.Now())
	if err != nil {
		writeQuarantineError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, stepQuarantineDTO{Quarantined: true, Quarantine: toQuarantineDTO(&q)})
}

// unquarantine answers DELETE on the same path; a step with no mark is fine.
func (h *quarantineHandler) unquarantine(w http.ResponseWriter, r *http.Request) {
	if !requireViaSession(w, r, "un-quarantine a step") {
		return
	}

	if err := h.service.UnquarantineStep(r.Context(), metrics.PipelineID(r.PathValue("pipelineId")), r.PathValue("step")); err != nil {
		writeQuarantineError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, stepQuarantineDTO{})
}

func writeQuarantineError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, metrics.ErrPipelineNotFound), errors.Is(err, metrics.ErrInvalidPipelineID):
		writeError(w, http.StatusNotFound, "not_found", "pipeline not found")
	case errors.Is(err, metrics.ErrQuarantineNoteTooLong):
		writeError(w, http.StatusUnprocessableEntity, "note_too_long", "a note is at most 500 characters")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "update quarantine")
	}
}

// readOptionalJSON decodes a JSON body into dst, treating an empty body as
// the zero value: a quarantine needs no note.
func readOptionalJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil && !errors.Is(err, io.EOF) {
		if isBodyTooLarge(err) {
			writePayloadTooLarge(w)
		} else {
			writeError(w, http.StatusBadRequest, "invalid_body", "malformed JSON body")
		}

		return false
	}

	return true
}
