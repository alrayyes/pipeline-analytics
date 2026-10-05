package httpserver

import (
	"net/http"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
)

type stageFailureDTO struct {
	Step     string  `json:"step"`
	Failures int     `json:"failures"`
	Share    float64 `json:"share"`
}

type categoryCountDTO struct {
	Category    string  `json:"category"`
	Occurrences int     `json:"occurrences"`
	Share       float64 `json:"share"`
}

type failingPipelineDTO struct {
	PipelineID   string `json:"pipelineId"`
	PipelineName string `json:"pipelineName"`
	RepoID       string `json:"repoId"`
	Runs         int    `json:"runs"`
	FailedRuns   int    `json:"failedRuns"`
}

type failureGroupPipelineDTO struct {
	PipelineID   string `json:"pipelineId"`
	PipelineName string `json:"pipelineName"`
}

type failureGroupDTO struct {
	Step        string                    `json:"step"`
	Category    string                    `json:"category"`
	Conclusion  string                    `json:"conclusion,omitempty"`
	Occurrences int                       `json:"occurrences"`
	Pipelines   []failureGroupPipelineDTO `json:"pipelines"`
}

// failureInsightsDTO is GET /api/insights/failures' response. The optional
// figures are pointers so an absent value is omitted instead of reading as
// zero, and every list is built non-nil so it encodes as [] rather than null.
type failureInsightsDTO struct {
	Window              string               `json:"window"`
	TotalRuns           int                  `json:"totalRuns"`
	FailedRuns          int                  `json:"failedRuns"`
	PassRate            *float64             `json:"passRate,omitempty"`
	PassRateDelta       *float64             `json:"passRateDelta,omitempty"`
	FlakyStepRatio      float64              `json:"flakyStepRatio"`
	MTTRSeconds         *float64             `json:"mttrSeconds,omitempty"`
	StageDistribution   []stageFailureDTO    `json:"stageDistribution"`
	CategoryBreakdown   []categoryCountDTO   `json:"categoryBreakdown"`
	TopFailingPipelines []failingPipelineDTO `json:"topFailingPipelines"`
	FailureGroups       []failureGroupDTO    `json:"failureGroups"`
}

func toFailureInsightsDTO(in metrics.FailureInsights) failureInsightsDTO {
	dto := failureInsightsDTO{
		Window:              in.Window.Label(),
		TotalRuns:           in.TotalRuns,
		FailedRuns:          in.FailedRuns,
		PassRate:            in.PassRate,
		PassRateDelta:       in.PassRateDelta,
		FlakyStepRatio:      in.FlakyStepRatio,
		StageDistribution:   make([]stageFailureDTO, 0, len(in.StageDistribution)),
		CategoryBreakdown:   make([]categoryCountDTO, 0, len(in.CategoryBreakdown)),
		TopFailingPipelines: make([]failingPipelineDTO, 0, len(in.TopFailingPipelines)),
		FailureGroups:       make([]failureGroupDTO, 0, len(in.FailureGroups)),
	}

	if in.MTTR != nil {
		seconds := in.MTTR.Seconds()
		dto.MTTRSeconds = &seconds
	}

	for _, s := range in.StageDistribution {
		dto.StageDistribution = append(dto.StageDistribution, stageFailureDTO{Step: s.Step, Failures: s.Failures, Share: s.Share})
	}

	for _, c := range in.CategoryBreakdown {
		dto.CategoryBreakdown = append(dto.CategoryBreakdown, categoryCountDTO{Category: string(c.Category), Occurrences: c.Occurrences, Share: c.Share})
	}

	for _, p := range in.TopFailingPipelines {
		dto.TopFailingPipelines = append(dto.TopFailingPipelines, failingPipelineDTO{
			PipelineID:   string(p.Pipeline.ID()),
			PipelineName: p.Pipeline.Name,
			RepoID:       p.Pipeline.RepoID,
			Runs:         p.Runs,
			FailedRuns:   p.FailedRuns,
		})
	}

	for _, g := range in.FailureGroups {
		pipelines := make([]failureGroupPipelineDTO, 0, len(g.Pipelines))
		for _, ref := range g.Pipelines {
			pipelines = append(pipelines, failureGroupPipelineDTO{PipelineID: string(ref.ID()), PipelineName: ref.Name})
		}

		dto.FailureGroups = append(dto.FailureGroups, failureGroupDTO{
			Step:        g.Step,
			Category:    string(g.Category),
			Conclusion:  g.Conclusion,
			Occurrences: g.Occurrences,
			Pipelines:   pipelines,
		})
	}

	return dto
}

type failureInsightsHandler struct {
	service *metrics.Service
}

// get answers GET /api/insights/failures: the failure overview's figures for
// a trailing window (24h, 7d or 30d; anything else is 7d, per the contract)
// and an optional repo and forge scope.
func (h *failureInsightsHandler) get(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	branch, ok := branchFilter(w, r)
	if !ok {
		return
	}

	insights, err := h.service.GetFailureInsights(
		r.Context(),
		time.Now(),
		metrics.ParseInsightWindow(query.Get("window")),
		metrics.InsightFilter{RepoID: query.Get("repoId"), Forge: query.Get("forge"), Branch: branch},
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "compute failure insights")

		return
	}

	writeJSON(w, http.StatusOK, toFailureInsightsDTO(insights))
}
