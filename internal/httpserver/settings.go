package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/settings"
)

type settingsDTO struct {
	Theme                 string `json:"theme"`
	ForgeFilter           string `json:"forgeFilter"`
	PipelinesHealthFilter string `json:"pipelinesHealthFilter"`
	PipelinesRepoSelector string `json:"pipelinesRepoSelector"`
	PipelinesSortOrder    string `json:"pipelinesSortOrder"`
}

func toSettingsDTO(s settings.Settings) settingsDTO {
	return settingsDTO{
		Theme:                 s.Theme,
		ForgeFilter:           s.ForgeFilter,
		PipelinesHealthFilter: s.PipelinesHealthFilter,
		PipelinesRepoSelector: s.PipelinesRepoSelector,
		PipelinesSortOrder:    s.PipelinesSortOrder,
	}
}

type settingsHandler struct {
	service *settings.Service
}

// get returns every setting, fully resolved against its documented
// default -- account-settings/spec.md's "Settings defaults".
func (h *settingsHandler) get(w http.ResponseWriter, r *http.Request) {
	info, ok := authInfoFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "no valid session or token")

		return
	}

	got, err := h.service.Get(r.Context(), info.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "get settings")

		return
	}

	writeJSON(w, http.StatusOK, toSettingsDTO(got))
}

// patch applies a partial update: a key present with a JSON value sets it,
// a key present with JSON null clears it back to its default, a key
// absent from the body is left untouched. Decoding into map[string]*string
// is what makes "present but null" distinguishable from "absent" --
// design.md's "null clears to default" convention.
func (h *settingsHandler) patch(w http.ResponseWriter, r *http.Request) {
	info, ok := authInfoFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "no valid session or token")

		return
	}

	var updates map[string]*string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "malformed JSON body")

		return
	}

	got, err := h.service.Update(r.Context(), info.UserID, updates)

	switch {
	case errors.Is(err, settings.ErrInvalidKey), errors.Is(err, settings.ErrInvalidValue):
		writeError(w, http.StatusBadRequest, "invalid_body", err.Error())
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal_error", "update settings")
	default:
		writeJSON(w, http.StatusOK, toSettingsDTO(got))
	}
}
