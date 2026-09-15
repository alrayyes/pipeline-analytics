// Package auth holds the dashboard-auth domain: the single dashboard user,
// their WebAuthn credentials, in-flight ceremonies, and sessions. See
// openspec/changes/add-pipeline-dashboard/specs/dashboard-auth/spec.md.
package auth

import (
	"context"
	"errors"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

// Sentinel errors, so a caller can errors.Is against a specific condition
// instead of matching on message text.
var (
	ErrUserExists       = errors.New("a user account already exists")
	ErrNoUser           = errors.New("no user account has been registered yet")
	ErrCeremonyNotFound = errors.New("ceremony not found or expired")
	ErrSessionNotFound  = errors.New("session not found or expired")
)

// User is the dashboard's single account.
type User struct {
	ID          string
	UserHandle  []byte
	DisplayName string
	CreatedAt   time.Time
}

// Store is the port the domain persists users, credentials, in-flight
// ceremonies, and sessions through.
type Store interface {
	// CreateUser creates the single dashboard user with the given WebAuthn
	// user handle (generated once, at the start of the registration
	// ceremony, and threaded through to here so the stored handle matches
	// the one the ceremony ran against). Returns ErrUserExists if a user
	// already exists.
	CreateUser(ctx context.Context, userHandle []byte, displayName string) (User, error)
	// GetUser returns the single dashboard user, or ErrNoUser if none has
	// registered yet.
	GetUser(ctx context.Context) (User, error)

	// PutCredential creates or updates (e.g. after a login's sign-count
	// bump) a credential belonging to userID.
	PutCredential(ctx context.Context, userID string, cred webauthn.Credential) error
	// CredentialsForUser returns every credential belonging to userID.
	CredentialsForUser(ctx context.Context, userID string) ([]webauthn.Credential, error)

	// SaveCeremony stores the SessionData for an in-flight WebAuthn
	// ceremony, keyed by an opaque id the caller generates.
	SaveCeremony(ctx context.Context, id string, data []byte, expiresAt time.Time) error
	// TakeCeremony reads and deletes a ceremony's data -- a ceremony is
	// single-use. Returns ErrCeremonyNotFound if it doesn't exist or has
	// expired.
	TakeCeremony(ctx context.Context, id string) ([]byte, error)

	// CreateSession creates a new session for userID, returning its opaque id.
	CreateSession(ctx context.Context, userID string) (string, error)
	// Session returns the userID a session id belongs to. Returns
	// ErrSessionNotFound if it doesn't exist or has expired.
	Session(ctx context.Context, sessionID string) (string, error)
	// DeleteSession ends a session.
	DeleteSession(ctx context.Context, sessionID string) error
}

// webauthnUser adapts a User and its credentials to webauthn.User.
type webauthnUser struct {
	user        User
	credentials []webauthn.Credential
}

func (u webauthnUser) WebAuthnID() []byte                         { return u.user.UserHandle }
func (u webauthnUser) WebAuthnName() string                       { return u.user.DisplayName }
func (u webauthnUser) WebAuthnDisplayName() string                { return u.user.DisplayName }
func (u webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }
