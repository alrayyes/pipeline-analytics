package httpserver

import (
	"context"
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
	tokens    ingestion.ForgeTokenStore
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

	writeJSON(w, http.StatusOK, toRepoListDTO(repos, hasMore))
}

// toRepoListDTO is the list endpoint's body, shared with the MCP list_repos
// tool so both report a repo the same way, token masked.
func toRepoListDTO(repos []ingestion.Repo, hasMore bool) repoListDTO {
	dtos := make([]repoDTO, 0, len(repos))
	for _, repo := range repos {
		dtos = append(dtos, toRepoDTO(repo))
	}

	return repoListDTO{Repos: dtos, HasMore: hasMore}
}

// isTrackedForge reports whether forge names one of the forges a repo can be
// tracked on, which the identifiers endpoint and its MCP tool both require.
func isTrackedForge(forge string) bool {
	return forge == string(ingestion.ForgeGitHub) || forge == string(ingestion.ForgeForgejo)
}

// apiProblem is a request that cannot be carried out: the status, code and
// message the endpoint answers with, which the MCP tools put in their error
// text, so an agent reads what a client of the endpoint reads.
type apiProblem struct {
	status        int
	code, message string
}

func (h *reposHandler) register(w http.ResponseWriter, r *http.Request) {
	var in repoRegistrationDTO
	if !readJSON(w, r, &in) {
		return
	}

	repo, problem := h.registerRepo(r.Context(), in)
	if problem != nil {
		writeError(w, problem.status, problem.code, problem.message)

		return
	}

	writeJSON(w, http.StatusCreated, toRepoDTO(repo))
}

// registerRepo is the registration, shared with the MCP register_repo tool:
// validate, find the token, register, and say why when it cannot.
func (h *reposHandler) registerRepo(ctx context.Context, in repoRegistrationDTO) (ingestion.Repo, *apiProblem) {
	if in.Identifier == "" || !validForge(in.Forge) {
		return ingestion.Repo{}, &apiProblem{http.StatusBadRequest, "invalid_body", "forge and identifier are required"}
	}

	if len(in.Identifier) > maxIdentifierLength || len(in.Token) > maxTokenLength || len(in.ForgejoInstanceURL) > maxInstanceURLLength {
		return ingestion.Repo{}, &apiProblem{http.StatusBadRequest, "invalid_body", "identifier, token or instance URL is too long"}
	}

	token, problem := givenOrSavedToken(ctx, h.tokens, in.Forge, in.ForgejoInstanceURL, in.Token)
	if problem != nil {
		return ingestion.Repo{}, problem
	}

	repo, err := h.registrar.Register(ctx, ingestion.NewRepo{
		Forge:              ingestion.Forge(in.Forge),
		Identifier:         in.Identifier,
		ForgejoInstanceURL: in.ForgejoInstanceURL,
		Token:              token,
	})
	if err != nil {
		if code, message, ok := registrationConflict(err); ok {
			slog.ErrorContext(ctx, "rejected repo registration",
				"forge", in.Forge, "identifier", in.Identifier, "reason", code)

			return ingestion.Repo{}, &apiProblem{http.StatusConflict, code, message}
		}

		slog.ErrorContext(ctx, "register repo failed",
			"forge", in.Forge, "identifier", in.Identifier, "error", err)

		return ingestion.Repo{}, &apiProblem{http.StatusInternalServerError, "internal_error", "register repo"}
	}

	return repo, nil
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
	if !isTrackedForge(forge) {
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
	if !readJSON(w, r, &in) {
		return
	}

	repos, problem := h.discoverRepos(r.Context(), in)
	if problem != nil {
		writeError(w, problem.status, problem.code, problem.message)

		return
	}

	writeJSON(w, http.StatusOK, repos)
}

// discoverRepos is the discovery, shared with the MCP discover_repos tool.
func (h *reposHandler) discoverRepos(ctx context.Context, in repoDiscoveryDTO) ([]string, *apiProblem) {
	if !validForge(in.Forge) {
		return nil, &apiProblem{http.StatusBadRequest, "invalid_body", "forge is required"}
	}

	if len(in.Token) > maxTokenLength || len(in.ForgejoInstanceURL) > maxInstanceURLLength {
		return nil, &apiProblem{http.StatusBadRequest, "invalid_body", "token or instance URL is too long"}
	}

	token, problem := givenOrSavedToken(ctx, h.tokens, in.Forge, in.ForgejoInstanceURL, in.Token)
	if problem != nil {
		return nil, problem
	}

	repos, err := h.registrar.Discover(ctx, ingestion.Forge(in.Forge), in.ForgejoInstanceURL, token)
	if err != nil {
		return nil, &apiProblem{http.StatusBadGateway, "forge_error", "could not list repositories: " + err.Error()}
	}

	return repos, nil
}

func (h *reposHandler) untrack(w http.ResponseWriter, r *http.Request) {
	if problem := h.untrackRepo(r.Context(), r.PathValue("repoId")); problem != nil {
		writeError(w, problem.status, problem.code, problem.message)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// untrackRepo is the removal, shared with the MCP untrack_repo tool.
func (h *reposHandler) untrackRepo(ctx context.Context, repoID string) *apiProblem {
	if err := h.store.DeleteRepo(ctx, repoID); err != nil {
		return &apiProblem{http.StatusNotFound, "not_found", "repo not found"}
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorDTO{Code: code, Message: message})
}
