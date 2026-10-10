package httpserver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
)

// flakyStepDTO is one entry of GET /api/steps/flaky.
type flakyStepDTO struct {
	PipelineID     string   `json:"pipelineId"`
	PipelineName   string   `json:"pipelineName"`
	RepoID         string   `json:"repoId"`
	Name           string   `json:"name"`
	FlakeRate      float64  `json:"flakeRate"`
	RunCount       int      `json:"runCount"`
	RecentOutcomes []string `json:"recentOutcomes"`
	// Quarantined is true when a person has marked this step as known; it is
	// still listed, with these figures unchanged.
	Quarantined bool           `json:"quarantined"`
	Quarantine  *quarantineDTO `json:"quarantine,omitempty"`
}

type flakyStepListDTO struct {
	Window  string         `json:"window"`
	Steps   []flakyStepDTO `json:"steps"`
	HasMore bool           `json:"hasMore"`
}

func toFlakyStepListDTO(window metrics.InsightWindow, steps []metrics.FlakyStep, hasMore bool) flakyStepListDTO {
	dtos := make([]flakyStepDTO, 0, len(steps))

	for _, s := range steps {
		outcomes := make([]string, 0, len(s.RecentOutcomes))
		for _, o := range s.RecentOutcomes {
			outcomes = append(outcomes, string(o))
		}

		dtos = append(dtos, flakyStepDTO{
			PipelineID:     string(s.Pipeline.ID()),
			PipelineName:   s.Pipeline.Name,
			RepoID:         s.Pipeline.RepoID,
			Name:           s.Name,
			FlakeRate:      s.FlakeRate,
			RunCount:       s.RunCount,
			RecentOutcomes: outcomes,
			Quarantined:    s.Quarantine != nil,
			Quarantine:     toQuarantineDTO(s.Quarantine),
		})
	}

	return flakyStepListDTO{Window: window.Label(), Steps: dtos, HasMore: hasMore}
}

type flakyStepsHandler struct {
	service *metrics.Service
}

// list answers GET /api/steps/flaky: the flaky steps across every pipeline
// for a trailing window (24h, 7d or 30d; anything else is 7d), ranked by
// flake rate, with an optional repo and forge scope.
func (h *flakyStepsHandler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	branch, ok := branchFilter(w, r)
	if !ok {
		return
	}

	window := metrics.ParseInsightWindow(query.Get("window"))
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))

	steps, hasMore, err := h.service.ListFlakySteps(
		r.Context(),
		time.Now(),
		window,
		metrics.InsightFilter{RepoID: query.Get("repoId"), Forge: query.Get("forge"), Branch: branch},
		limit,
		offset,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list flaky steps")

		return
	}

	writeJSON(w, http.StatusOK, toFlakyStepListDTO(window, steps, hasMore))
}
