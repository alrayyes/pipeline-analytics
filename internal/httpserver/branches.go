package httpserver

import (
	"net/http"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
)

type branchEntryDTO struct {
	Name     string `json:"name"`
	RunCount int    `json:"runCount"`
}

type branchListDTO struct {
	Window   string           `json:"window"`
	Branches []branchEntryDTO `json:"branches"`
}

func toBranchListDTO(window metrics.InsightWindow, branches []metrics.BranchCount) branchListDTO {
	entries := make([]branchEntryDTO, 0, len(branches))
	for _, b := range branches {
		entries = append(entries, branchEntryDTO{Name: b.Name, RunCount: b.RunCount})
	}

	return branchListDTO{Window: window.Label(), Branches: entries}
}

type branchesHandler struct {
	service *metrics.Service
}

// list answers GET /api/branches: the branches with runs in a trailing
// window (24h, 7d or 30d; anything else is 7d), busiest first, with an
// optional repo and forge scope.
func (h *branchesHandler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	window := metrics.ParseInsightWindow(query.Get("window"))

	branches, err := h.service.ListBranches(
		r.Context(),
		time.Now(),
		window,
		metrics.InsightFilter{RepoID: query.Get("repoId"), Forge: query.Get("forge")},
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list branches")

		return
	}

	writeJSON(w, http.StatusOK, toBranchListDTO(window, branches))
}
