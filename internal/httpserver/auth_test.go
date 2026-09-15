package httpserver_test

import (
	"context"
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
