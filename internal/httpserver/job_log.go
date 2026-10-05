package httpserver

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

type jobLogDTO struct {
	Available bool     `json:"available"`
	Reason    string   `json:"reason,omitempty"`
	Lines     []string `json:"lines"`
	Truncated bool     `json:"truncated"`
	ForgeURL  string   `json:"forgeUrl"`
}

func toJobLogDTO(r ingestion.JobLogResult) jobLogDTO {
	lines := r.Lines
	if lines == nil {
		lines = []string{}
	}

	return jobLogDTO{
		Available: r.Available,
		Reason:    string(r.Reason),
		Lines:     lines,
		Truncated: r.Truncated,
		ForgeURL:  r.ForgeURL,
	}
}

type jobLogHandler struct {
	service *ingestion.JobLogService
}

// get returns the tail of one job's log, fetched from the forge when asked
// and never stored. A log the forge can't supply is a 200 with
// available=false and a reason, so the dashboard can say why and still link
// out; only an unknown run or job is a 404 (openapi.yaml, getJobLog).
func (h *jobLogHandler) get(w http.ResponseWriter, r *http.Request) {
	lines := 0

	if raw := r.URL.Query().Get("lines"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > ingestion.MaxLogLines {
			writeError(w, http.StatusBadRequest, "invalid_lines", "lines must be a whole number from 1 to 1000")

			return
		}

		lines = n
	}

	result, err := h.service.Tail(r.Context(), r.PathValue("runId"), r.PathValue("jobId"), lines)
	if err != nil {
		if errors.Is(err, ingestion.ErrJobNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "job not found")

			return
		}

		writeError(w, http.StatusInternalServerError, "internal_error", "get job log")

		return
	}

	writeJSON(w, http.StatusOK, toJobLogDTO(result))
}
