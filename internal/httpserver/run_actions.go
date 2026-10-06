package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

type runActionsHandler struct {
	service *ingestion.RunActionService
}

// rerun asks the forge to re-run a concluded run (openapi.yaml, rerunRun).
func (h *runActionsHandler) rerun(w http.ResponseWriter, r *http.Request) {
	h.do(w, r, h.service.Rerun)
}

// cancel asks the forge to cancel a run that hasn't concluded (openapi.yaml,
// cancelRun).
func (h *runActionsHandler) cancel(w http.ResponseWriter, r *http.Request) {
	h.do(w, r, h.service.Cancel)
}

// do runs one action. Session-only: an API token gets a 401, because this
// writes to the forge and an agent doesn't get that.
func (h *runActionsHandler) do(w http.ResponseWriter, r *http.Request, act func(context.Context, string) error) {
	info, ok := authInfoFromContext(r)
	if !ok || !info.ViaSession {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "a session is required to act on a run")

		return
	}

	err := act(r.Context(), r.PathValue("runId"))

	switch {
	case err == nil:
		w.WriteHeader(http.StatusAccepted)
	case errors.Is(err, ingestion.ErrRunNotFound):
		writeError(w, http.StatusNotFound, "not_found", "run not found")
	case errors.Is(err, ingestion.ErrNotActionable):
		writeError(w, http.StatusConflict, "not_actionable", "the run's state doesn't allow this")
	case errors.Is(err, ingestion.ErrActionForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "the saved token needs Actions write permission to do this")
	case errors.Is(err, ingestion.ErrActionUnsupported):
		writeError(w, http.StatusNotImplemented, "unsupported", "this forge has no API for this action")
	case errors.Is(err, ingestion.ErrActionUnreachable):
		writeError(w, http.StatusBadGateway, "unreachable", "the forge didn't answer")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "act on run")
	}
}
