package httpserver

import (
	"errors"
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/settings"
)

type settingsValuesDTO struct {
	Theme                 string `json:"theme"`
	ForgeFilter           string `json:"forgeFilter"`
	PipelinesHealthFilter string `json:"pipelinesHealthFilter"`
	PipelinesRepoSelector string `json:"pipelinesRepoSelector"`
	PipelinesSortOrder    string `json:"pipelinesSortOrder"`
	TelemetryWindow       string `json:"telemetryWindow"`
}

// settingsDTO is the settings in force and, beside them, the server's
// defaults, so a client never keeps a copy of those.
type settingsDTO struct {
	settingsValuesDTO

	Defaults settingsValuesDTO `json:"defaults"`
}

func toSettingsDTO(s settings.Settings) settingsDTO {
	return settingsDTO{settingsValuesDTO: toSettingsValuesDTO(s), Defaults: toSettingsValuesDTO(settings.Defaults())}
}

func toSettingsValuesDTO(s settings.Settings) settingsValuesDTO {
	return settingsValuesDTO{
		Theme:                 s.Theme,
		ForgeFilter:           s.ForgeFilter,
		PipelinesHealthFilter: s.PipelinesHealthFilter,
		PipelinesRepoSelector: s.PipelinesRepoSelector,
		PipelinesSortOrder:    s.PipelinesSortOrder,
		TelemetryWindow:       s.TelemetryWindow,
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
	if !readJSON(w, r, &updates) {
		return
	}

	got, err := h.service.Update(r.Context(), info.UserID, updates)
	if err != nil {
		code, message := classifySettingsError(err)
		status := http.StatusInternalServerError

		if code == "invalid_body" {
			status = http.StatusBadRequest
		}

		writeError(w, status, code, message)

		return
	}

	writeJSON(w, http.StatusOK, toSettingsDTO(got))
}

// classifySettingsError is how a failed update is reported: the code and
// message the endpoint answers with, which the MCP tool puts in its error text.
// An unknown key or a value outside its set is the caller's mistake (400
// invalid_body); anything else is ours.
func classifySettingsError(err error) (code, message string) {
	if errors.Is(err, settings.ErrInvalidKey) || errors.Is(err, settings.ErrInvalidValue) {
		return "invalid_body", err.Error()
	}

	return "internal_error", "update settings"
}
