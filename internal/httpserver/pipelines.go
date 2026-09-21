package httpserver

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
)

type pipelineSummaryDTO struct {
	ID               string     `json:"id"`
	RepoID           string     `json:"repoId"`
	Name             string     `json:"name"`
	HealthStatus     string     `json:"healthStatus"`
	TriggeredSignals []string   `json:"triggeredSignals,omitempty"`
	LastRunAt        *time.Time `json:"lastRunAt,omitempty"`
}

func toPipelineSummaryDTO(p metrics.Pipeline) pipelineSummaryDTO {
	signals := make([]string, 0, len(p.TriggeredSignals))
	for _, s := range p.TriggeredSignals {
		signals = append(signals, string(s))
	}

	return pipelineSummaryDTO{
		ID:               string(p.ID),
		RepoID:           p.RepoID,
		Name:             p.Name,
		HealthStatus:     string(p.HealthStatus),
		TriggeredSignals: signals,
		LastRunAt:        p.LastRunAt,
	}
}

type pipelineListDTO struct {
	Pipelines []pipelineSummaryDTO `json:"pipelines"`
	HasMore   bool                 `json:"hasMore"`
}

type trendDTO struct {
	Timestamps []time.Time `json:"timestamps"`
	P50        []float64   `json:"p50,omitempty"`
	P90        []float64   `json:"p90,omitempty"`
	Rate       []float64   `json:"rate,omitempty"`
}

func toTrendDTO(t metrics.Trend) trendDTO {
	timestamps := t.Timestamps
	if timestamps == nil {
		timestamps = []time.Time{}
	}

	return trendDTO{Timestamps: timestamps, P50: t.P50, P90: t.P90, Rate: t.Rate}
}

type pipelineDetailDTO struct {
	pipelineSummaryDTO
	DurationTrend    trendDTO `json:"durationTrend"`
	FailureRateTrend trendDTO `json:"failureRateTrend"`
}

type stepDTO struct {
	ID                          string  `json:"id"`
	Name                        string  `json:"name"`
	DurationContributionSeconds float64 `json:"durationContributionSeconds"`
	QueueSeconds                float64 `json:"queueSeconds"`
	ExecSeconds                 float64 `json:"execSeconds"`
	FailureRate                 float64 `json:"failureRate"`
	FailureCount                int     `json:"failureCount"`
	Flaky                       bool    `json:"flaky"`
	ForgeURL                    string  `json:"forgeUrl,omitempty"`
}

func toStepDTO(s metrics.Step) stepDTO {
	return stepDTO{
		ID:                          s.ID,
		Name:                        s.Name,
		DurationContributionSeconds: s.DurationContributionSeconds,
		QueueSeconds:                s.QueueSeconds,
		ExecSeconds:                 s.ExecSeconds,
		FailureRate:                 s.FailureRate,
		FailureCount:                s.FailureCount,
		Flaky:                       s.Flaky,
		ForgeURL:                    s.ForgeURL,
	}
}

type flakyRunDTO struct {
	RunID     string     `json:"runId"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
	ForgeURL  string     `json:"forgeUrl"`
}

func toFlakyRunDTO(r metrics.FlakyRun) flakyRunDTO {
	return flakyRunDTO{RunID: r.RunID, StartedAt: r.StartedAt, ForgeURL: r.ForgeURL}
}

type runStepDTO struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion,omitempty"`
	ForgeURL   string `json:"forgeUrl,omitempty"`
}

func toRunStepDTO(s metrics.RunStep) runStepDTO {
	return runStepDTO{Name: s.Name, Status: s.Status, Conclusion: s.Conclusion, ForgeURL: s.ForgeURL}
}

type runDetailDTO struct {
	RunID     string       `json:"runId"`
	StartedAt *time.Time   `json:"startedAt,omitempty"`
	Steps     []runStepDTO `json:"steps"`
}

type pipelinesHandler struct {
	service *metrics.Service
}

func (h *pipelinesHandler) list(w http.ResponseWriter, r *http.Request) {
	window := metrics.ParseWindow(r.URL.Query().Get("window"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	pipelines, hasMore, err := h.service.ListPipelines(r.Context(), window, metrics.PipelineListFilter{
		RepoID: r.URL.Query().Get("repoId"),
		Forge:  r.URL.Query().Get("forge"),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list pipelines")

		return
	}

	dtos := make([]pipelineSummaryDTO, 0, len(pipelines))
	for _, p := range pipelines {
		dtos = append(dtos, toPipelineSummaryDTO(p))
	}

	writeJSON(w, http.StatusOK, pipelineListDTO{Pipelines: dtos, HasMore: hasMore})
}

func (h *pipelinesHandler) get(w http.ResponseWriter, r *http.Request) {
	window := metrics.ParseWindow(r.URL.Query().Get("window"))

	detail, err := h.service.GetPipeline(r.Context(), metrics.PipelineID(r.PathValue("pipelineId")), window)
	if err != nil {
		if errors.Is(err, metrics.ErrPipelineNotFound) || errors.Is(err, metrics.ErrInvalidPipelineID) {
			writeError(w, http.StatusNotFound, "not_found", "pipeline not found")

			return
		}

		writeError(w, http.StatusInternalServerError, "internal_error", "get pipeline")

		return
	}

	writeJSON(w, http.StatusOK, pipelineDetailDTO{
		pipelineSummaryDTO: toPipelineSummaryDTO(detail.Pipeline),
		DurationTrend:      toTrendDTO(detail.DurationTrend),
		FailureRateTrend:   toTrendDTO(detail.FailureRateTrend),
	})
}

// steps returns a pipeline's step ranking. Unlike get, an unknown
// pipelineId isn't distinguished from one with no recorded steps yet --
// both report an empty list, since this is a collection endpoint rather
// than a single-resource lookup. Only a malformed id 404s.
func (h *pipelinesHandler) steps(w http.ResponseWriter, r *http.Request) {
	window := metrics.ParseWindow(r.URL.Query().Get("window"))

	steps, err := h.service.GetPipelineSteps(r.Context(), metrics.PipelineID(r.PathValue("pipelineId")), window)
	if err != nil {
		if errors.Is(err, metrics.ErrInvalidPipelineID) {
			writeError(w, http.StatusNotFound, "not_found", "pipeline not found")

			return
		}

		writeError(w, http.StatusInternalServerError, "internal_error", "get pipeline steps")

		return
	}

	dtos := make([]stepDTO, 0, len(steps))
	for _, s := range steps {
		dtos = append(dtos, toStepDTO(s))
	}

	writeJSON(w, http.StatusOK, dtos)
}

// flakyRuns returns every run in which the step named by the "step" query
// parameter failed -- the step-scoped drill-down GetPipelineSteps' Steps
// table links a flaky step into, rather than an arbitrary occurrence.
func (h *pipelinesHandler) flakyRuns(w http.ResponseWriter, r *http.Request) {
	window := metrics.ParseWindow(r.URL.Query().Get("window"))

	runs, err := h.service.ListFlakyRuns(r.Context(), metrics.PipelineID(r.PathValue("pipelineId")), r.URL.Query().Get("step"), window)
	if err != nil {
		if errors.Is(err, metrics.ErrInvalidPipelineID) {
			writeError(w, http.StatusNotFound, "not_found", "pipeline not found")

			return
		}

		writeError(w, http.StatusInternalServerError, "internal_error", "list flaky runs")

		return
	}

	dtos := make([]flakyRunDTO, 0, len(runs))
	for _, run := range runs {
		dtos = append(dtos, toFlakyRunDTO(run))
	}

	writeJSON(w, http.StatusOK, dtos)
}

// runSteps returns one run's own steps and statuses -- what a flaky run
// drills down into, so its forgeUrl points at the exact job that ran.
func (h *pipelinesHandler) runSteps(w http.ResponseWriter, r *http.Request) {
	detail, err := h.service.GetRunSteps(r.Context(), r.PathValue("runId"))
	if err != nil {
		if errors.Is(err, metrics.ErrRunNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "run not found")

			return
		}

		writeError(w, http.StatusInternalServerError, "internal_error", "get run steps")

		return
	}

	steps := make([]runStepDTO, 0, len(detail.Steps))
	for _, s := range detail.Steps {
		steps = append(steps, toRunStepDTO(s))
	}

	writeJSON(w, http.StatusOK, runDetailDTO{RunID: detail.RunID, StartedAt: detail.StartedAt, Steps: steps})
}

type pipelineStepsGroupDTO struct {
	PipelineID   string    `json:"pipelineId"`
	PipelineName string    `json:"pipelineName"`
	RepoID       string    `json:"repoId"`
	Steps        []stepDTO `json:"steps"`
}

type pipelineStepsGroupListDTO struct {
	Groups  []pipelineStepsGroupDTO `json:"groups"`
	HasMore bool                    `json:"hasMore"`
}

// unhealthySteps returns a page of pipelines with a flaky or failing step,
// grouped by pipeline -- the cross-pipeline counterpart to steps, which is
// scoped to one pipeline.
func (h *pipelinesHandler) unhealthySteps(w http.ResponseWriter, r *http.Request) {
	window := metrics.ParseWindow(r.URL.Query().Get("window"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	groups, hasMore, err := h.service.ListUnhealthySteps(r.Context(), window, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list unhealthy steps")

		return
	}

	dtos := make([]pipelineStepsGroupDTO, 0, len(groups))

	for _, g := range groups {
		steps := make([]stepDTO, 0, len(g.Steps))
		for _, s := range g.Steps {
			steps = append(steps, toStepDTO(s))
		}

		dtos = append(dtos, pipelineStepsGroupDTO{
			PipelineID:   string(g.PipelineID),
			PipelineName: g.PipelineName,
			RepoID:       g.RepoID,
			Steps:        steps,
		})
	}

	writeJSON(w, http.StatusOK, pipelineStepsGroupListDTO{Groups: dtos, HasMore: hasMore})
}

type usageEntryDTO struct {
	Workflow      string  `json:"workflow"`
	RunnerMinutes float64 `json:"runnerMinutes"`
}

type usageHandler struct {
	service *metrics.Service
	repos   ingestion.Store
}

func (h *usageHandler) get(w http.ResponseWriter, r *http.Request) {
	repoID := r.PathValue("repoId")

	if _, err := h.repos.GetRepo(r.Context(), repoID); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "repo not found")

		return
	}

	window := metrics.ParseWindow(r.URL.Query().Get("window"))

	usage, err := h.service.GetRepoUsage(r.Context(), repoID, window)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "get repo usage")

		return
	}

	dtos := make([]usageEntryDTO, 0, len(usage))
	for _, u := range usage {
		dtos = append(dtos, usageEntryDTO{Workflow: u.Workflow, RunnerMinutes: u.RunnerMinutes})
	}

	writeJSON(w, http.StatusOK, dtos)
}
