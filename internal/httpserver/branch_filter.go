package httpserver

import (
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
)

// branchFilter reads the optional `branch` query parameter, answering 400
// itself when it is longer than the spec allows.
func branchFilter(w http.ResponseWriter, r *http.Request) (string, bool) {
	branch := r.URL.Query().Get("branch")
	if len(branch) > metrics.MaxBranchLength {
		writeError(w, http.StatusBadRequest, "invalid_branch", "branch must be at most 255 characters")

		return "", false
	}

	return branch, true
}
