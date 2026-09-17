package httpserver

import (
	"errors"
	"net/http"
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
		Flaky:                       s.Flaky,
		ForgeURL:                    s.ForgeURL,
	}
}

type pipelinesHandler struct {
	service *metrics.Service
}

func (h *pipelinesHandler) list(w http.ResponseWriter, r *http.Request) {
	window := metrics.ParseWindow(r.URL.Query().Get("window"))

	pipelines, err := h.service.ListPipelines(r.Context(), window)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list pipelines")

		return
	}

	dtos := make([]pipelineSummaryDTO, 0, len(pipelines))
	for _, p := range pipelines {
		dtos = append(dtos, toPipelineSummaryDTO(p))
	}

	writeJSON(w, http.StatusOK, dtos)
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
