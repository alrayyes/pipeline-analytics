package httpserver

import (
	"errors"
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
)

type savedForgeTokenDTO struct {
	ID                 string `json:"id"`
	Forge              string `json:"forge"`
	ForgejoInstanceURL string `json:"forgejoInstanceUrl,omitempty"`
	TokenMasked        string `json:"tokenMasked"`
}

func toSavedForgeTokenDTO(t ingestion.SavedToken) savedForgeTokenDTO {
	return savedForgeTokenDTO{
		ID:                 t.ID,
		Forge:              string(t.Forge),
		ForgejoInstanceURL: t.ForgejoInstanceURL,
		TokenMasked:        t.TokenMasked,
	}
}

type saveForgeTokenDTO struct {
	Forge              string `json:"forge"`
	ForgejoInstanceURL string `json:"forgejoInstanceUrl"`
	Token              string `json:"token"`
}

// forgeTokensHandler manages the tokens saved for registering repos. All
// three routes are session-only, like API tokens and passkeys, and none ever
// returns a token: only its masked form (openapi.yaml, /api/forge-tokens).
type forgeTokensHandler struct {
	store ingestion.ForgeTokenStore
}

// requireSessionAuth answers 401 unless the request came from a session.
func requireSessionAuth(w http.ResponseWriter, r *http.Request) bool {
	return requireViaSession(w, r, "manage saved tokens")
}

// requireViaSession answers 401 unless the request came from a session, saying
// what the session is needed to do.
func requireViaSession(w http.ResponseWriter, r *http.Request, to string) bool {
	info, ok := authInfoFromContext(r)
	if !ok || !info.ViaSession {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "a session is required to "+to)

		return false
	}

	return true
}

func (h *forgeTokensHandler) list(w http.ResponseWriter, r *http.Request) {
	if !requireSessionAuth(w, r) {
		return
	}

	tokens, err := h.store.ListForgeTokens(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list saved tokens")

		return
	}

	dtos := make([]savedForgeTokenDTO, 0, len(tokens))
	for _, t := range tokens {
		dtos = append(dtos, toSavedForgeTokenDTO(t))
	}

	writeJSON(w, http.StatusOK, map[string][]savedForgeTokenDTO{"tokens": dtos})
}

func (h *forgeTokensHandler) save(w http.ResponseWriter, r *http.Request) {
	if !requireSessionAuth(w, r) {
		return
	}

	var in saveForgeTokenDTO
	if !readJSON(w, r, &in) {
		return
	}

	if !validForge(in.Forge) || in.Token == "" || (in.Forge == string(ingestion.ForgeForgejo) && in.ForgejoInstanceURL == "") {
		writeError(w, http.StatusBadRequest, "invalid_body", "forge and token are required, and a Forgejo token needs its instance URL")

		return
	}

	if len(in.Token) > maxTokenLength || len(in.ForgejoInstanceURL) > maxInstanceURLLength {
		writeError(w, http.StatusBadRequest, "invalid_body", "token or instance URL is too long")

		return
	}

	saved, err := h.store.SaveForgeToken(r.Context(), ingestion.Forge(in.Forge), in.ForgejoInstanceURL, in.Token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "save token")

		return
	}

	writeJSON(w, http.StatusOK, toSavedForgeTokenDTO(saved))
}

func (h *forgeTokensHandler) delete(w http.ResponseWriter, r *http.Request) {
	if !requireSessionAuth(w, r) {
		return
	}

	err := h.store.DeleteForgeToken(r.Context(), r.PathValue("tokenId"))
	if errors.Is(err, ingestion.ErrSavedTokenNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "saved token not found")

		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "delete saved token")

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validForge(forge string) bool {
	return forge == string(ingestion.ForgeGitHub) || forge == string(ingestion.ForgeForgejo)
}

// resolveToken is the token a registration or discovery request should use:
// the one it carries, else the one saved for its forge and instance. When
// neither exists it answers 400 no_saved_token itself.
func resolveToken(w http.ResponseWriter, r *http.Request, tokens ingestion.ForgeTokenStore, forge, instanceURL, given string) (string, bool) {
	if given != "" {
		return given, true
	}

	if tokens != nil {
		token, err := tokens.ForgeToken(r.Context(), ingestion.Forge(forge), instanceURL)
		if err == nil {
			return token, true
		}

		if !errors.Is(err, ingestion.ErrSavedTokenNotFound) {
			writeError(w, http.StatusInternalServerError, "internal_error", "read saved token")

			return "", false
		}
	}

	writeError(w, http.StatusBadRequest, "no_saved_token", "no token given and none saved for this forge")

	return "", false
}
