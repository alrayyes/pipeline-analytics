package httpserver

import (
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
)

const (
	ceremonyCookieName = "pa_ceremony"
	sessionCookieName  = "session"
)

type authHandler struct {
	service *auth.Service
}

func (h *authHandler) registerOptions(w http.ResponseWriter, r *http.Request) {
	creation, ceremonyID, err := h.service.BeginRegistration(r.Context())

	switch {
	case errors.Is(err, auth.ErrUserExists):
		writeError(w, http.StatusConflict, "user_exists", "a user account already exists")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal_error", "begin registration")
	default:
		setCookie(w, ceremonyCookieName, ceremonyID, 5*time.Minute)
		writeJSON(w, http.StatusOK, creation)
	}
}

func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	ceremonyID, ok := ceremonyCookie(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_body", "no registration in progress")

		return
	}

	user, sessionID, err := h.service.FinishRegistration(r.Context(), ceremonyID, r)

	switch {
	case errors.Is(err, auth.ErrUserExists):
		writeError(w, http.StatusConflict, "user_exists", "a user account already exists")
	case err != nil:
		writeError(w, http.StatusBadRequest, "invalid_body", "registration failed")
	default:
		clearCookie(w, ceremonyCookieName)
		setCookie(w, sessionCookieName, sessionID, 0)
		writeJSON(w, http.StatusCreated, toUserDTO(user))
	}
}

func (h *authHandler) loginOptions(w http.ResponseWriter, r *http.Request) {
	assertion, ceremonyID, err := h.service.BeginLogin(r.Context())

	switch {
	case errors.Is(err, auth.ErrNoUser):
		writeError(w, http.StatusNotFound, "no_user", "no account has been registered yet")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal_error", "begin login")
	default:
		setCookie(w, ceremonyCookieName, ceremonyID, 5*time.Minute)
		writeJSON(w, http.StatusOK, assertion)
	}
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	ceremonyID, ok := ceremonyCookie(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "no login in progress")

		return
	}

	sessionID, err := h.service.FinishLogin(r.Context(), ceremonyID, r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "login failed")

		return
	}

	clearCookie(w, ceremonyCookieName)
	setCookie(w, sessionCookieName, sessionID, 0)
	w.WriteHeader(http.StatusOK)
}

func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		_ = h.service.Logout(r.Context(), cookie.Value)
	}

	clearCookie(w, sessionCookieName)
	w.WriteHeader(http.StatusNoContent)
}

// issueToken creates a new API token. Session-only: see
// add-api-token-auth/design.md's "session-only" decision -- a token can't
// mint another token.
func (h *authHandler) issueToken(w http.ResponseWriter, r *http.Request) {
	info, ok := authInfoFromContext(r)
	if !ok || !info.ViaSession {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "a session is required to issue a token")

		return
	}

	tok, raw, err := h.service.IssueToken(r.Context(), info.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "issue token")

		return
	}

	writeJSON(w, http.StatusCreated, toTokenDTO(tok, raw))
}

// revokeToken revokes an API token by id. Session-only, same reasoning as
// issueToken.
func (h *authHandler) revokeToken(w http.ResponseWriter, r *http.Request) {
	info, ok := authInfoFromContext(r)
	if !ok || !info.ViaSession {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "a session is required to revoke a token")

		return
	}

	err := h.service.RevokeToken(r.Context(), info.UserID, r.PathValue("tokenId"))

	switch {
	case errors.Is(err, auth.ErrTokenNotFound):
		writeError(w, http.StatusNotFound, "not_found", "token not found")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal_error", "revoke token")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

// addCredentialOptions starts the authenticated "add another passkey"
// ceremony. Session-only, matching issueToken/revokeToken's reasoning: an
// API token shouldn't be able to enroll another credential on the account
// any more than it can mint another token.
func (h *authHandler) addCredentialOptions(w http.ResponseWriter, r *http.Request) {
	info, ok := authInfoFromContext(r)
	if !ok || !info.ViaSession {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "a session is required to add a credential")

		return
	}

	creation, ceremonyID, err := h.service.BeginAddCredential(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "begin add credential")

		return
	}

	setCookie(w, ceremonyCookieName, ceremonyID, 5*time.Minute)
	writeJSON(w, http.StatusOK, creation)
}

// addCredential completes the ceremony started by addCredentialOptions,
// storing the new credential under the caller-supplied label.
func (h *authHandler) addCredential(w http.ResponseWriter, r *http.Request) {
	info, ok := authInfoFromContext(r)
	if !ok || !info.ViaSession {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "a session is required to add a credential")

		return
	}

	ceremonyID, ok := ceremonyCookie(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_body", "no add-credential ceremony in progress")

		return
	}

	label := r.URL.Query().Get("label")

	err := h.service.FinishAddCredential(r.Context(), ceremonyID, label, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "add credential failed")

		return
	}

	clearCookie(w, ceremonyCookieName)
	w.WriteHeader(http.StatusCreated)
}

// listCredentials lists the account's registered credentials. Session-only,
// same reasoning as addCredentialOptions.
func (h *authHandler) listCredentials(w http.ResponseWriter, r *http.Request) {
	info, ok := authInfoFromContext(r)
	if !ok || !info.ViaSession {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "a session is required to list credentials")

		return
	}

	infos, err := h.service.ListCredentials(r.Context(), info.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "list credentials")

		return
	}

	writeJSON(w, http.StatusOK, toCredentialDTOs(infos))
}

// revokeCredential revokes a credential by its base64url-encoded id.
// Session-only, same reasoning as addCredentialOptions.
func (h *authHandler) revokeCredential(w http.ResponseWriter, r *http.Request) {
	info, ok := authInfoFromContext(r)
	if !ok || !info.ViaSession {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "a session is required to revoke a credential")

		return
	}

	credentialID, err := base64.RawURLEncoding.DecodeString(r.PathValue("credentialId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "malformed credential id")

		return
	}

	err = h.service.RevokeCredential(r.Context(), info.UserID, credentialID)

	switch {
	case errors.Is(err, auth.ErrLastCredential):
		writeError(w, http.StatusConflict, "last_credential", "cannot revoke the account's last remaining credential")
	case errors.Is(err, auth.ErrCredentialNotFound):
		writeError(w, http.StatusNotFound, "not_found", "credential not found")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal_error", "revoke credential")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

type userDTO struct {
	DisplayName string `json:"displayName"`
}

func toUserDTO(u auth.User) userDTO {
	return userDTO{DisplayName: u.DisplayName}
}

// tokenDTO is only ever built right after issuance, at the one point the
// raw value exists -- see add-api-token-auth/design.md's "Hash the token
// at rest" decision.
type tokenDTO struct {
	ID        string    `json:"id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func toTokenDTO(t auth.Token, raw string) tokenDTO {
	return tokenDTO{ID: t.ID, Token: raw, CreatedAt: t.CreatedAt, ExpiresAt: t.ExpiresAt}
}

// credentialDTO is a registered credential's listing metadata -- the
// credential id is base64url-encoded for the wire (support-multiple-
// passkeys/design.md's "Credential identifier over the wire" decision),
// matching the same encoding already used for other WebAuthn binary fields
// in this codebase's ceremony JSON.
type credentialDTO struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"createdAt"`
}

func toCredentialDTOs(infos []auth.CredentialInfo) []credentialDTO {
	dtos := make([]credentialDTO, len(infos))
	for i, info := range infos {
		dtos[i] = credentialDTO{
			ID:        base64.RawURLEncoding.EncodeToString(info.ID),
			Label:     info.Label,
			CreatedAt: info.CreatedAt,
		}
	}

	return dtos
}

func ceremonyCookie(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(ceremonyCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}

	return cookie.Value, true
}

// setCookie sets an HttpOnly, Secure, SameSite=Strict cookie. maxAge of 0
// means a session cookie (cleared when the browser closes) rather than no
// expiry -- appropriate for the short-lived ceremony cookie; the real
// session cookie relies on the server-side session's own TTL instead of a
// client-visible expiry.
func setCookie(w http.ResponseWriter, name, value string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}
