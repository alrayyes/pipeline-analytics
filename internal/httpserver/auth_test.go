package httpserver_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	authsqlite "github.com/alrayyes/pipeline-analytics/internal/auth/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/alrayyes/pipeline-analytics/internal/httpserver"
	ingestionsqlite "github.com/alrayyes/pipeline-analytics/internal/ingestion/sqlite"
	"github.com/descope/virtualwebauthn"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/require"
)

const (
	authTestRPID   = "example.com"
	authTestOrigin = "https://example.com"
)

func newAuthTestServer(t *testing.T) http.Handler {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	authStore := authsqlite.NewStore(conn)
	ingestionStore := ingestionsqlite.NewStore(conn, make([]byte, 32))

	web, err := gowebauthn.New(&gowebauthn.Config{
		RPID:          authTestRPID,
		RPDisplayName: "pipeline-analytics",
		RPOrigins:     []string{authTestOrigin},
	})
	require.NoError(t, err)

	return httpserver.New(httpserver.Deps{
		Auth:           auth.NewService(web, authStore),
		AuthStore:      authStore,
		IngestionStore: ingestionStore,
		RunStore:       ingestionStore,
		Version:        "test-version",
		Assets:         fstest.MapFS{"index.html": {Data: []byte("<html></html>")}},
	})
}

// TestAuthHTTPFlow drives the full registration and login ceremonies
// through the real HTTP handlers and cookies -- not the Service directly --
// against a simulated FIDO2 authenticator.
func TestAuthHTTPFlow(t *testing.T) {
	t.Parallel()

	srv := newAuthTestServer(t)
	rp := virtualwebauthn.RelyingParty{Name: "pipeline-analytics", ID: authTestRPID, Origin: authTestOrigin}
	authenticator := virtualwebauthn.NewAuthenticator()
	credential := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)

	var sessionCookie *http.Cookie

	t.Run("register", func(t *testing.T) {
		optsReq := httptest.NewRequest(http.MethodPost, "/api/auth/register/options", nil)
		optsRec := httptest.NewRecorder()
		srv.ServeHTTP(optsRec, optsReq)
		require.Equal(t, http.StatusOK, optsRec.Code)

		ceremonyCookie := findCookie(t, optsRec, "pa_ceremony")

		attestationOptions, err := virtualwebauthn.ParseAttestationOptions(optsRec.Body.String())
		require.NoError(t, err)

		attestationResponse := virtualwebauthn.CreateAttestationResponse(rp, authenticator, credential, *attestationOptions)

		finishReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(attestationResponse))
		finishReq.AddCookie(ceremonyCookie)
		finishRec := httptest.NewRecorder()
		srv.ServeHTTP(finishRec, finishReq)
		require.Equal(t, http.StatusCreated, finishRec.Code)

		authenticator.Options.UserHandle = []byte(attestationOptions.UserID)
		authenticator.AddCredential(credential)
	})

	t.Run("registering a second account is rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register/options", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		require.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("login", func(t *testing.T) {
		optsReq := httptest.NewRequest(http.MethodPost, "/api/auth/login/options", nil)
		optsRec := httptest.NewRecorder()
		srv.ServeHTTP(optsRec, optsReq)
		require.Equal(t, http.StatusOK, optsRec.Code)

		ceremonyCookie := findCookie(t, optsRec, "pa_ceremony")

		assertionOptions, err := virtualwebauthn.ParseAssertionOptions(optsRec.Body.String())
		require.NoError(t, err)

		assertionResponse := virtualwebauthn.CreateAssertionResponse(rp, authenticator, credential, *assertionOptions)

		finishReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(assertionResponse))
		finishReq.AddCookie(ceremonyCookie)
		finishRec := httptest.NewRecorder()
		srv.ServeHTTP(finishRec, finishReq)
		require.Equal(t, http.StatusOK, finishRec.Code)

		sessionCookie = findCookie(t, finishRec, "session")
	})

	t.Run("logout clears the session", func(t *testing.T) {
		// Confirm the session actually authenticates a gated endpoint first.
		before := httptest.NewRequest(http.MethodGet, "/api/repos", nil)
		before.AddCookie(sessionCookie)
		beforeRec := httptest.NewRecorder()
		srv.ServeHTTP(beforeRec, before)
		require.NotEqual(t, http.StatusUnauthorized, beforeRec.Code)

		logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		logoutReq.AddCookie(sessionCookie)
		logoutRec := httptest.NewRecorder()
		srv.ServeHTTP(logoutRec, logoutReq)
		require.Equal(t, http.StatusNoContent, logoutRec.Code)

		// The same (now-deleted) session cookie no longer authenticates a
		// gated endpoint.
		after := httptest.NewRequest(http.MethodGet, "/api/repos", nil)
		after.AddCookie(sessionCookie)
		afterRec := httptest.NewRecorder()
		srv.ServeHTTP(afterRec, after)
		require.Equal(t, http.StatusUnauthorized, afterRec.Code)
	})
}

// registerAndLogin drives the full WebAuthn ceremony against srv and
// returns a valid session cookie -- shared setup for tests that need an
// authenticated session but aren't testing the ceremony itself.
func registerAndLogin(t *testing.T, srv http.Handler) *http.Cookie {
	t.Helper()

	rp := virtualwebauthn.RelyingParty{Name: "pipeline-analytics", ID: authTestRPID, Origin: authTestOrigin}
	authenticator := virtualwebauthn.NewAuthenticator()
	credential := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)

	optsReq := httptest.NewRequest(http.MethodPost, "/api/auth/register/options", nil)
	optsRec := httptest.NewRecorder()
	srv.ServeHTTP(optsRec, optsReq)
	require.Equal(t, http.StatusOK, optsRec.Code)

	ceremonyCookie := findCookie(t, optsRec, "pa_ceremony")

	attestationOptions, err := virtualwebauthn.ParseAttestationOptions(optsRec.Body.String())
	require.NoError(t, err)

	attestationResponse := virtualwebauthn.CreateAttestationResponse(rp, authenticator, credential, *attestationOptions)

	finishReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(attestationResponse))
	finishReq.AddCookie(ceremonyCookie)
	finishRec := httptest.NewRecorder()
	srv.ServeHTTP(finishRec, finishReq)
	require.Equal(t, http.StatusCreated, finishRec.Code)

	return findCookie(t, finishRec, "session")
}

// TestAPITokenHTTPFlow drives issuance, bearer-authenticated use, and
// revocation of an API token through the real HTTP handlers -- per #178's
// acceptance criteria.
func TestAPITokenHTTPFlow(t *testing.T) {
	t.Parallel()

	srv := newAuthTestServer(t)
	sessionCookie := registerAndLogin(t, srv)

	t.Run("issuing a token without a session is denied", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/tokens", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("issuing, using, and revoking a token", func(t *testing.T) {
		t.Parallel()

		issueReq := httptest.NewRequest(http.MethodPost, "/api/auth/tokens", nil)
		issueReq.AddCookie(sessionCookie)
		issueRec := httptest.NewRecorder()
		srv.ServeHTTP(issueRec, issueReq)
		require.Equal(t, http.StatusCreated, issueRec.Code)

		var issued struct {
			ID    string `json:"id"`
			Token string `json:"token"`
		}
		require.NoError(t, json.NewDecoder(issueRec.Body).Decode(&issued))
		require.NotEmpty(t, issued.ID)
		require.NotEmpty(t, issued.Token)

		// The raw token authenticates a gated endpoint with no session
		// cookie at all.
		useReq := httptest.NewRequest(http.MethodGet, "/api/repos", nil)
		useReq.Header.Set("Authorization", "Bearer "+issued.Token)
		useRec := httptest.NewRecorder()
		srv.ServeHTTP(useRec, useReq)
		require.NotEqual(t, http.StatusUnauthorized, useRec.Code)

		// A bearer token alone can't issue or revoke tokens.
		bearerIssueReq := httptest.NewRequest(http.MethodPost, "/api/auth/tokens", nil)
		bearerIssueReq.Header.Set("Authorization", "Bearer "+issued.Token)
		bearerIssueRec := httptest.NewRecorder()
		srv.ServeHTTP(bearerIssueRec, bearerIssueReq)
		require.Equal(t, http.StatusUnauthorized, bearerIssueRec.Code)

		bearerRevokeReq := httptest.NewRequest(http.MethodDelete, "/api/auth/tokens/"+issued.ID, nil)
		bearerRevokeReq.Header.Set("Authorization", "Bearer "+issued.Token)
		bearerRevokeRec := httptest.NewRecorder()
		srv.ServeHTTP(bearerRevokeRec, bearerRevokeReq)
		require.Equal(t, http.StatusUnauthorized, bearerRevokeRec.Code)

		// Revoking with the session works, and the token stops working.
		revokeReq := httptest.NewRequest(http.MethodDelete, "/api/auth/tokens/"+issued.ID, nil)
		revokeReq.AddCookie(sessionCookie)
		revokeRec := httptest.NewRecorder()
		srv.ServeHTTP(revokeRec, revokeReq)
		require.Equal(t, http.StatusNoContent, revokeRec.Code)

		afterReq := httptest.NewRequest(http.MethodGet, "/api/repos", nil)
		afterReq.Header.Set("Authorization", "Bearer "+issued.Token)
		afterRec := httptest.NewRecorder()
		srv.ServeHTTP(afterRec, afterReq)
		require.Equal(t, http.StatusUnauthorized, afterRec.Code)

		var errBody struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		require.NoError(t, json.NewDecoder(afterRec.Body).Decode(&errBody))
		require.NotEmpty(t, errBody.Code)
		require.NotEmpty(t, errBody.Message)
	})
}

// TestCredentialManagementHTTPFlow drives the authenticated add/list/
// revoke credential endpoints through the real HTTP handlers -- per #188's
// acceptance criteria.
func TestCredentialManagementHTTPFlow(t *testing.T) {
	t.Parallel()

	srv := newAuthTestServer(t)
	sessionCookie := registerAndLogin(t, srv)
	rp := virtualwebauthn.RelyingParty{Name: "pipeline-analytics", ID: authTestRPID, Origin: authTestOrigin}

	t.Run("adding a credential without a session is denied", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/credentials/options", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("listing credentials without a session is denied", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/api/auth/credentials", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("adding, listing, and revoking a credential", func(t *testing.T) {
		t.Parallel()

		optsReq := httptest.NewRequest(http.MethodPost, "/api/auth/credentials/options", nil)
		optsReq.AddCookie(sessionCookie)
		optsRec := httptest.NewRecorder()
		srv.ServeHTTP(optsRec, optsReq)
		require.Equal(t, http.StatusOK, optsRec.Code)

		addCeremonyCookie := findCookie(t, optsRec, "pa_ceremony")

		attestationOptions, err := virtualwebauthn.ParseAttestationOptions(optsRec.Body.String())
		require.NoError(t, err)

		secondAuthenticator := virtualwebauthn.NewAuthenticator()
		secondCredential := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)
		attestationResponse := virtualwebauthn.CreateAttestationResponse(rp, secondAuthenticator, secondCredential, *attestationOptions)

		finishReq := httptest.NewRequest(http.MethodPost, "/api/auth/credentials?label=MacBook", strings.NewReader(attestationResponse))
		finishReq.AddCookie(sessionCookie)
		finishReq.AddCookie(addCeremonyCookie)
		finishRec := httptest.NewRecorder()
		srv.ServeHTTP(finishRec, finishReq)
		require.Equal(t, http.StatusCreated, finishRec.Code)

		// A bearer token can't enroll a credential -- session-only, same as
		// token issuance/revocation.
		bearerReq := httptest.NewRequest(http.MethodPost, "/api/auth/credentials/options", nil)
		bearerReq.Header.Set("Authorization", "Bearer not-a-real-token")
		bearerRec := httptest.NewRecorder()
		srv.ServeHTTP(bearerRec, bearerReq)
		require.Equal(t, http.StatusUnauthorized, bearerRec.Code)

		listReq := httptest.NewRequest(http.MethodGet, "/api/auth/credentials", nil)
		listReq.AddCookie(sessionCookie)
		listRec := httptest.NewRecorder()
		srv.ServeHTTP(listRec, listReq)
		require.Equal(t, http.StatusOK, listRec.Code)

		var listed []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		}
		require.NoError(t, json.NewDecoder(listRec.Body).Decode(&listed))
		require.Len(t, listed, 2)
		require.Empty(t, listed[0].Label)
		require.Equal(t, "MacBook", listed[1].Label)

		// Checked while both credentials still exist -- once only one is
		// left, any revoke (even of an unknown id) is rejected as the
		// last-credential case first.
		unknownID := base64.RawURLEncoding.EncodeToString([]byte("nonexistent"))
		unknownReq := httptest.NewRequest(http.MethodDelete, "/api/auth/credentials/"+unknownID, nil)
		unknownReq.AddCookie(sessionCookie)
		unknownRec := httptest.NewRecorder()
		srv.ServeHTTP(unknownRec, unknownReq)
		require.Equal(t, http.StatusNotFound, unknownRec.Code)

		malformedReq := httptest.NewRequest(http.MethodDelete, "/api/auth/credentials/not-valid-base64!!!", nil)
		malformedReq.AddCookie(sessionCookie)
		malformedRec := httptest.NewRecorder()
		srv.ServeHTTP(malformedRec, malformedReq)
		require.Equal(t, http.StatusBadRequest, malformedRec.Code)

		revokeFirstReq := httptest.NewRequest(http.MethodDelete, "/api/auth/credentials/"+listed[0].ID, nil)
		revokeFirstReq.AddCookie(sessionCookie)
		revokeFirstRec := httptest.NewRecorder()
		srv.ServeHTTP(revokeFirstRec, revokeFirstReq)
		require.Equal(t, http.StatusNoContent, revokeFirstRec.Code)

		// The account's last remaining credential can't be revoked.
		revokeLastReq := httptest.NewRequest(http.MethodDelete, "/api/auth/credentials/"+listed[1].ID, nil)
		revokeLastReq.AddCookie(sessionCookie)
		revokeLastRec := httptest.NewRecorder()
		srv.ServeHTTP(revokeLastRec, revokeLastReq)
		require.Equal(t, http.StatusConflict, revokeLastRec.Code)
	})
}

func TestRequireSession(t *testing.T) {
	t.Parallel()

	publicPaths := []string{"/healthz", "/api/version", "/api/auth/register/options", "/webhooks/github", "/", "/releases"}
	for _, path := range publicPaths {
		t.Run("public: "+path, func(t *testing.T) {
			t.Parallel()

			srv := newAuthTestServer(t)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)

			require.NotEqual(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

func findCookie(t *testing.T, rec *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()

	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}

	t.Fatalf("no %q cookie in response", name)

	return nil
}
