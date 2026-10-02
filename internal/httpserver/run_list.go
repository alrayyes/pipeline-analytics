package httpserver

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
)

// runSummaryDTO is one entry of GET /api/runs. What isn't known (a run still
// going has no conclusion or duration, an older run has no commit fields) is
// omitted rather than sent as an empty string or zero, so a client can tell
// "no data" from a real value.
type runSummaryDTO struct {
	ID              string       `json:"id"`
	PipelineID      string       `json:"pipelineId"`
	PipelineName    string       `json:"pipelineName"`
	RepoID          string       `json:"repoId"`
	Status          string       `json:"status"`
	Conclusion      string       `json:"conclusion,omitempty"`
	Outcome         string       `json:"outcome"`
	StartedAt       *time.Time   `json:"startedAt,omitempty"`
	DurationSeconds *float64     `json:"durationSeconds,omitempty"`
	Branch          string       `json:"branch,omitempty"`
	SHA             string       `json:"sha,omitempty"`
	Message         string       `json:"message,omitempty"`
	Actor           string       `json:"actor,omitempty"`
	ForgeURL        string       `json:"forgeUrl,omitempty"`
	Steps           []runStepDTO `json:"steps"`
}

type runListDTO struct {
	Runs    []runSummaryDTO `json:"runs"`
	HasMore bool            `json:"hasMore"`
}

func toRunSummaryDTO(e metrics.RunEntry) runSummaryDTO {
	dto := runSummaryDTO{
		ID:           e.ID,
		PipelineID:   string(e.Pipeline.ID()),
		PipelineName: e.Pipeline.Name,
		RepoID:       e.Pipeline.RepoID,
		Status:       e.Status,
		Conclusion:   e.Conclusion,
		Outcome:      string(metrics.OutcomeOf(e.Status, e.Conclusion)),
		StartedAt:    e.StartedAt,
		Branch:       e.Branch,
		SHA:          e.SHA,
		Message:      e.Message,
		Actor:        e.Actor,
		ForgeURL:     e.ForgeURL,
		Steps:        make([]runStepDTO, 0, len(e.Steps)),
	}

	if secs, ok := e.DurationSeconds(); ok {
		dto.DurationSeconds = &secs
	}

	for _, step := range e.Steps {
		dto.Steps = append(dto.Steps, toRunStepDTO(step))
	}

	return dto
}

type runListHandler struct {
	service *metrics.Service
}

// list answers GET /api/runs: a page of runs, newest first, each with its
// steps for a stage progression bar. An unknown status is a 400 rather than
// an empty list, so a typo can't read as "no failures".
func (h *runListHandler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	status, err := metrics.ParseRunStatus(query.Get("status"))
	if errors.Is(err, metrics.ErrInvalidRunStatus) {
		writeError(w, http.StatusBadRequest, "invalid_status", "status must be one of all, failed, running, success")

		return
	}

	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	runs, hasMore, err := h.service.ListRuns(r.Context(), metrics.RunListFilter{
		RepoID: query.Get("repoId"),
		Forge:  query.Get("forge"),
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list runs")

		return
	}

	dtos := make([]runSummaryDTO, 0, len(runs))
	for _, run := range runs {
		dtos = append(dtos, toRunSummaryDTO(run))
	}

	writeJSON(w, http.StatusOK, runListDTO{Runs: dtos, HasMore: hasMore})
}
