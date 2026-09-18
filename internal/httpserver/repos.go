package httpserver

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

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

type repoListDTO struct {
	Repos   []repoDTO `json:"repos"`
	HasMore bool      `json:"hasMore"`
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
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	repos, hasMore, err := h.store.ListRepos(r.Context(), ingestion.RepoListFilter{
		Forge:  ingestion.Forge(r.URL.Query().Get("forge")),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list repos")

		return
	}

	dtos := make([]repoDTO, 0, len(repos))
	for _, repo := range repos {
		dtos = append(dtos, toRepoDTO(repo))
	}

	writeJSON(w, http.StatusOK, repoListDTO{Repos: dtos, HasMore: hasMore})
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
		if code, message, ok := registrationConflict(err); ok {
			slog.ErrorContext(r.Context(), "rejected repo registration",
				"forge", in.Forge, "identifier", in.Identifier, "reason", code)
			writeError(w, http.StatusConflict, code, message)

			return
		}

		slog.ErrorContext(r.Context(), "register repo failed",
			"forge", in.Forge, "identifier", in.Identifier, "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "register repo")

		return
	}

	writeJSON(w, http.StatusCreated, toRepoDTO(repo))
}

// registrationConflict maps a Registrar.Register error to the (code,
// message) pair register's 409 response uses, if err is one of the
// registration-rejection sentinels -- already-tracked, or archived/fork/
// mirror (exclude-forge-mirrors-from-registration). Every other error stays
// a 500.
func registrationConflict(err error) (code, message string, ok bool) {
	switch {
	case errors.Is(err, ingestion.ErrRepoAlreadyTracked):
		return "already_tracked", "this repo is already tracked", true
	case errors.Is(err, ingestion.ErrRepoArchived):
		return "repo_archived", "this repo is archived", true
	case errors.Is(err, ingestion.ErrRepoFork):
		return "repo_fork", "this repo is a fork", true
	case errors.Is(err, ingestion.ErrRepoMirror):
		return "repo_mirror", "this repo is a mirror", true
	default:
		return "", "", false
	}
}

type repoIdentifiersDTO struct {
	Identifiers []string `json:"identifiers"`
}

func (h *reposHandler) identifiers(w http.ResponseWriter, r *http.Request) {
	forge := r.URL.Query().Get("forge")
	if forge != string(ingestion.ForgeGitHub) && forge != string(ingestion.ForgeForgejo) {
		writeError(w, http.StatusBadRequest, "invalid_query", "forge is required")

		return
	}

	identifiers, err := h.store.ListRepoIdentifiers(r.Context(), ingestion.Forge(forge), r.URL.Query().Get("forgejoInstanceUrl"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list repo identifiers")

		return
	}

	writeJSON(w, http.StatusOK, repoIdentifiersDTO{Identifiers: identifiers})
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
