package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

type repoDTO struct {
	ID                    string `json:"id"`
	Forge                 string `json:"forge"`
	Identifier            string `json:"identifier"`
	ForgejoInstanceURL    string `json:"forgejoInstanceUrl,omitempty"`
	TokenMasked           string `json:"tokenMasked"`
	IngestionStatus       string `json:"ingestionStatus"`
	IngestionStatusReason string `json:"ingestionStatusReason,omitempty"`
}

type repoRegistrationDTO struct {
	Forge              string `json:"forge"`
	Identifier         string `json:"identifier"`
	ForgejoInstanceURL string `json:"forgejoInstanceUrl,omitempty"`
	Token              string `json:"token"`
}

type repoDiscoveryDTO struct {
	Forge              string `json:"forge"`
	ForgejoInstanceURL string `json:"forgejoInstanceUrl,omitempty"`
	Token              string `json:"token"`
}

type errorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func toRepoDTO(r ingestion.Repo) repoDTO {
	return repoDTO{
		ID:                    r.ID,
		Forge:                 string(r.Forge),
		Identifier:            r.Identifier,
		ForgejoInstanceURL:    r.ForgejoInstanceURL,
		TokenMasked:           r.TokenMasked,
		IngestionStatus:       string(r.IngestionStatus),
		IngestionStatusReason: r.IngestionStatusReason,
	}
}

type reposHandler struct {
	registrar *ingestion.Registrar
	store     ingestion.Store
}

func (h *reposHandler) list(w http.ResponseWriter, r *http.Request) {
	repos, err := h.store.ListRepos(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list repos")

		return
	}

	dtos := make([]repoDTO, 0, len(repos))
	for _, repo := range repos {
		dtos = append(dtos, toRepoDTO(repo))
	}

	writeJSON(w, http.StatusOK, dtos)
}

func (h *reposHandler) register(w http.ResponseWriter, r *http.Request) {
	var in repoRegistrationDTO
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "malformed JSON body")

		return
	}

	if in.Identifier == "" || in.Token == "" || (in.Forge != string(ingestion.ForgeGitHub) && in.Forge != string(ingestion.ForgeForgejo)) {
		writeError(w, http.StatusBadRequest, "invalid_body", "forge, identifier, and token are required")

		return
	}

	repo, err := h.registrar.Register(r.Context(), ingestion.NewRepo{
		Forge:              ingestion.Forge(in.Forge),
		Identifier:         in.Identifier,
		ForgejoInstanceURL: in.ForgejoInstanceURL,
		Token:              in.Token,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "register repo")

		return
	}

	writeJSON(w, http.StatusCreated, toRepoDTO(repo))
}

func (h *reposHandler) discover(w http.ResponseWriter, r *http.Request) {
	var in repoDiscoveryDTO
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "malformed JSON body")

		return
	}

	if in.Token == "" || (in.Forge != string(ingestion.ForgeGitHub) && in.Forge != string(ingestion.ForgeForgejo)) {
		writeError(w, http.StatusBadRequest, "invalid_body", "forge and token are required")

		return
	}

	repos, err := h.registrar.Discover(r.Context(), ingestion.Forge(in.Forge), in.ForgejoInstanceURL, in.Token)
	if err != nil {
		writeError(w, http.StatusBadGateway, "forge_error", "could not list repositories: "+err.Error())

		return
	}

	writeJSON(w, http.StatusOK, repos)
}

func (h *reposHandler) untrack(w http.ResponseWriter, r *http.Request) {
	if err := h.store.DeleteRepo(r.Context(), r.PathValue("repoId")); err != nil {
		writeError(w, http.StatusNotFound, "not_found", "repo not found")

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorDTO{Code: code, Message: message})
}
