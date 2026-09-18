package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/auth"
	"github.com/alrayyes/pipeline-analytics/internal/auth/sqlite"
	"github.com/alrayyes/pipeline-analytics/internal/db"
	"github.com/descope/virtualwebauthn"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/require"
)

const (
	testRPID   = "localhost"
	testOrigin = "https://localhost"
)

func newTestService(t *testing.T) *auth.Service {
	t.Helper()

	conn, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })
	require.NoError(t, db.Migrate(context.Background(), conn))

	store := sqlite.NewStore(conn)

	web, err := gowebauthn.New(&gowebauthn.Config{
		RPID:          testRPID,
		RPDisplayName: "pipeline-analytics",
		RPOrigins:     []string{testOrigin},
	})
	require.NoError(t, err)

	return auth.NewService(web, store)
}

// TestService_FullCeremony exercises real registration then login against a
// simulated FIDO2 authenticator (github.com/descope/virtualwebauthn), rather
// than hand-stubbing the cryptographic responses.
func TestService_FullCeremony(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service := newTestService(t)
	rp := virtualwebauthn.RelyingParty{Name: "pipeline-analytics", ID: testRPID, Origin: testOrigin}
	authenticator := virtualwebauthn.NewAuthenticator()
	credential := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)

	t.Run("register", func(t *testing.T) {
		creation, ceremonyID, err := service.BeginRegistration(ctx)
		require.NoError(t, err)

		creationJSON, err := json.Marshal(creation)
		require.NoError(t, err)

		attestationOptions, err := virtualwebauthn.ParseAttestationOptions(string(creationJSON))
		require.NoError(t, err)

		attestationResponse := virtualwebauthn.CreateAttestationResponse(rp, authenticator, credential, *attestationOptions)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(attestationResponse))
		req.Header.Set("Content-Type", "application/json")

		user, sessionID, err := service.FinishRegistration(ctx, ceremonyID, req)
		require.NoError(t, err)
		require.Equal(t, "admin", user.DisplayName)
		require.NotEmpty(t, sessionID)

		authenticator.Options.UserHandle = []byte(attestationOptions.UserID)
		authenticator.AddCredential(credential)
	})

	t.Run("registering a second account is rejected", func(t *testing.T) {
		_, _, err := service.BeginRegistration(ctx)
		require.ErrorIs(t, err, auth.ErrUserExists)
	})

	t.Run("login", func(t *testing.T) {
		assertion, ceremonyID, err := service.BeginLogin(ctx)
		require.NoError(t, err)

		assertionJSON, err := json.Marshal(assertion)
		require.NoError(t, err)

		assertionOptions, err := virtualwebauthn.ParseAssertionOptions(string(assertionJSON))
		require.NoError(t, err)

		assertionResponse := virtualwebauthn.CreateAssertionResponse(rp, authenticator, credential, *assertionOptions)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(assertionResponse))
		req.Header.Set("Content-Type", "application/json")

		sessionID, err := service.FinishLogin(ctx, ceremonyID, req)
		require.NoError(t, err)
		require.NotEmpty(t, sessionID)
	})

	t.Run("a failed assertion does not establish a session", func(t *testing.T) {
		assertion, ceremonyID, err := service.BeginLogin(ctx)
		require.NoError(t, err)

		assertionJSON, err := json.Marshal(assertion)
		require.NoError(t, err)

		assertionOptions, err := virtualwebauthn.ParseAssertionOptions(string(assertionJSON))
		require.NoError(t, err)

		// A credential the authenticator never registered against this RP.
		wrongCredential := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)
		wrongAuthenticator := virtualwebauthn.NewAuthenticator()
		wrongAuthenticator.Options.UserHandle = authenticator.Options.UserHandle
		wrongAuthenticator.AddCredential(wrongCredential)

		assertionResponse := virtualwebauthn.CreateAssertionResponse(rp, wrongAuthenticator, wrongCredential, *assertionOptions)

		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(assertionResponse))
		req.Header.Set("Content-Type", "application/json")

		_, err = service.FinishLogin(ctx, ceremonyID, req)
		require.Error(t, err)
	})
}

func TestService_Token(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	service := newTestService(t)

	// The token store is keyed by user id -- create a real one via
	// registration rather than an arbitrary string, so this exercises the
	// same path a real issuance would.
	creation, ceremonyID, err := service.BeginRegistration(ctx)
	require.NoError(t, err)

	creationJSON, err := json.Marshal(creation)
	require.NoError(t, err)

	rp := virtualwebauthn.RelyingParty{Name: "pipeline-analytics", ID: testRPID, Origin: testOrigin}
	authenticator := virtualwebauthn.NewAuthenticator()
	credential := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)

	attestationOptions, err := virtualwebauthn.ParseAttestationOptions(string(creationJSON))
	require.NoError(t, err)

	attestationResponse := virtualwebauthn.CreateAttestationResponse(rp, authenticator, credential, *attestationOptions)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(attestationResponse))
	req.Header.Set("Content-Type", "application/json")

	user, _, err := service.FinishRegistration(ctx, ceremonyID, req)
	require.NoError(t, err)

	t.Run("issues a token that authenticates, then revokes it", func(t *testing.T) {
		tok, raw, err := service.IssueToken(ctx, user.ID)
		require.NoError(t, err)
		require.NotEmpty(t, raw)

		gotUserID, err := service.AuthenticateToken(ctx, raw)
		require.NoError(t, err)
		require.Equal(t, user.ID, gotUserID)

		require.NoError(t, service.RevokeToken(ctx, user.ID, tok.ID))

		_, err = service.AuthenticateToken(ctx, raw)
		require.ErrorIs(t, err, auth.ErrTokenNotFound)
	})

	t.Run("an unknown raw token does not authenticate", func(t *testing.T) {
		_, err := service.AuthenticateToken(ctx, "not-a-real-token")
		require.ErrorIs(t, err, auth.ErrTokenNotFound)
	})
}

func TestService_BeginLogin_NoUser(t *testing.T) {
	t.Parallel()

	service := newTestService(t)
	_, _, err := service.BeginLogin(context.Background())
	require.ErrorIs(t, err, auth.ErrNoUser)
}

func TestService_FinishRegistration_UnknownCeremony(t *testing.T) {
	t.Parallel()

	service := newTestService(t)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader("{}"))
	_, _, err := service.FinishRegistration(context.Background(), "does-not-exist", req)
	require.ErrorIs(t, err, auth.ErrCeremonyNotFound)
}

func TestService_FinishLogin_UnknownCeremony(t *testing.T) {
	t.Parallel()

	service := newTestService(t)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("{}"))
	_, err := service.FinishLogin(context.Background(), "does-not-exist", req)
	require.ErrorIs(t, err, auth.ErrCeremonyNotFound)
}
