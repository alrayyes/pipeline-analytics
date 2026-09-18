package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/crypto"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// ceremonyTTL bounds how long a WebAuthn ceremony (the window between
// fetching options and completing the browser/authenticator round trip)
// stays valid.
const ceremonyTTL = 5 * time.Minute

// registrationDisplayName is the single dashboard user's display name.
// There's exactly one account and no UI asks for one, so it's fixed.
const registrationDisplayName = "admin"

// Service orchestrates the WebAuthn registration and login ceremonies:
// begin, persist the in-flight ceremony state, finish, and issue a session.
type Service struct {
	webAuthn *webauthn.WebAuthn
	store    Store
}

// NewService returns a Service.
func NewService(webAuthn *webauthn.WebAuthn, store Store) *Service {
	return &Service{webAuthn: webAuthn, store: store}
}

// BeginRegistration starts the registration ceremony. Returns ErrUserExists
// if the single dashboard account already exists.
func (s *Service) BeginRegistration(ctx context.Context) (*protocol.CredentialCreation, string, error) {
	if _, err := s.store.GetUser(ctx); err == nil {
		return nil, "", ErrUserExists
	} else if !errors.Is(err, ErrNoUser) {
		return nil, "", fmt.Errorf("check existing user: %w", err)
	}

	handle := make([]byte, 64)
	if _, err := rand.Read(handle); err != nil {
		return nil, "", fmt.Errorf("generate user handle: %w", err)
	}

	tempUser := webauthnUser{user: User{UserHandle: handle, DisplayName: registrationDisplayName}}

	creation, session, err := s.webAuthn.BeginRegistration(tempUser)
	if err != nil {
		return nil, "", fmt.Errorf("begin registration: %w", err)
	}

	ceremonyID, err := s.saveCeremony(ctx, session)
	if err != nil {
		return nil, "", err
	}

	return creation, ceremonyID, nil
}

// FinishRegistration completes the registration ceremony started by
// BeginRegistration, creating the dashboard user, storing their credential,
// and establishing a session (the OpenAPI contract's "account created,
// session established" -- registration logs the new account straight in,
// the same as a login would).
func (s *Service) FinishRegistration(ctx context.Context, ceremonyID string, r *http.Request) (User, string, error) {
	session, err := s.loadCeremony(ctx, ceremonyID)
	if err != nil {
		return User{}, "", err
	}

	tempUser := webauthnUser{user: User{UserHandle: session.UserID, DisplayName: registrationDisplayName}}

	cred, err := s.webAuthn.FinishRegistration(tempUser, *session, r)
	if err != nil {
		return User{}, "", fmt.Errorf("finish registration: %w", err)
	}

	user, err := s.store.CreateUser(ctx, session.UserID, registrationDisplayName)
	if err != nil {
		return User{}, "", fmt.Errorf("create user: %w", err)
	}

	if err := s.store.PutCredential(ctx, user.ID, *cred); err != nil {
		return User{}, "", fmt.Errorf("save credential: %w", err)
	}

	sessionID, err := s.store.CreateSession(ctx, user.ID)
	if err != nil {
		return User{}, "", fmt.Errorf("create session: %w", err)
	}

	return user, sessionID, nil
}

// BeginLogin starts the login ceremony. Returns ErrNoUser if no dashboard
// account has been registered yet.
func (s *Service) BeginLogin(ctx context.Context) (*protocol.CredentialAssertion, string, error) {
	user, creds, err := s.userWithCredentials(ctx)
	if err != nil {
		return nil, "", err
	}

	assertion, session, err := s.webAuthn.BeginLogin(webauthnUser{user: user, credentials: creds})
	if err != nil {
		return nil, "", fmt.Errorf("begin login: %w", err)
	}

	ceremonyID, err := s.saveCeremony(ctx, session)
	if err != nil {
		return nil, "", err
	}

	return assertion, ceremonyID, nil
}

// FinishLogin completes the login ceremony started by BeginLogin, returning
// a new session id on success.
func (s *Service) FinishLogin(ctx context.Context, ceremonyID string, r *http.Request) (string, error) {
	session, err := s.loadCeremony(ctx, ceremonyID)
	if err != nil {
		return "", err
	}

	user, creds, err := s.userWithCredentials(ctx)
	if err != nil {
		return "", err
	}

	cred, err := s.webAuthn.FinishLogin(webauthnUser{user: user, credentials: creds}, *session, r)
	if err != nil {
		return "", fmt.Errorf("finish login: %w", err)
	}

	if err := s.store.PutCredential(ctx, user.ID, *cred); err != nil {
		return "", fmt.Errorf("save credential: %w", err)
	}

	sessionID, err := s.store.CreateSession(ctx, user.ID)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return sessionID, nil
}

// Logout ends a session.
func (s *Service) Logout(ctx context.Context, sessionID string) error {
	if err := s.store.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

// IssueToken creates a new API token for userID, returning its metadata
// and its raw secret -- the only time the raw secret is ever available.
func (s *Service) IssueToken(ctx context.Context, userID string) (Token, string, error) {
	tok, raw, err := s.store.CreateToken(ctx, userID)
	if err != nil {
		return Token{}, "", fmt.Errorf("create token: %w", err)
	}

	return tok, raw, nil
}

// AuthenticateToken returns the userID a raw API token belongs to. Returns
// ErrTokenNotFound if the token doesn't exist, has expired, or has been
// revoked.
func (s *Service) AuthenticateToken(ctx context.Context, rawToken string) (string, error) {
	userID, err := s.store.TokenUserID(ctx, rawToken)
	if err != nil {
		return "", fmt.Errorf("look up token: %w", err)
	}

	return userID, nil
}

// RevokeToken revokes tokenID, scoped to userID. Returns ErrTokenNotFound
// if no such token exists for that user.
func (s *Service) RevokeToken(ctx context.Context, userID, tokenID string) error {
	if err := s.store.RevokeToken(ctx, userID, tokenID); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}

	return nil
}

func (s *Service) userWithCredentials(ctx context.Context) (User, []webauthn.Credential, error) {
	user, err := s.store.GetUser(ctx)
	if err != nil {
		return User{}, nil, fmt.Errorf("load user: %w", err)
	}

	creds, err := s.store.CredentialsForUser(ctx, user.ID)
	if err != nil {
		return User{}, nil, fmt.Errorf("load credentials: %w", err)
	}

	return user, creds, nil
}

func (s *Service) saveCeremony(ctx context.Context, session *webauthn.SessionData) (string, error) {
	id, err := crypto.RandomHex(32)
	if err != nil {
		return "", fmt.Errorf("generate ceremony id: %w", err)
	}

	data, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("marshal ceremony: %w", err)
	}

	if err := s.store.SaveCeremony(ctx, id, data, time.Now().Add(ceremonyTTL)); err != nil {
		return "", fmt.Errorf("save ceremony: %w", err)
	}

	return id, nil
}

func (s *Service) loadCeremony(ctx context.Context, id string) (*webauthn.SessionData, error) {
	data, err := s.store.TakeCeremony(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load ceremony: %w", err)
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("unmarshal ceremony: %w", err)
	}

	return &session, nil
}
