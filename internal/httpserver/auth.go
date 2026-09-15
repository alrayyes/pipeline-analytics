package httpserver

import (
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

type userDTO struct {
	DisplayName string `json:"displayName"`
}

func toUserDTO(u auth.User) userDTO {
	return userDTO{DisplayName: u.DisplayName}
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
