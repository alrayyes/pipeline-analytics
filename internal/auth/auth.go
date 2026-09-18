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
	ErrUserExists         = errors.New("a user account already exists")
	ErrNoUser             = errors.New("no user account has been registered yet")
	ErrCeremonyNotFound   = errors.New("ceremony not found or expired")
	ErrSessionNotFound    = errors.New("session not found or expired")
	ErrTokenNotFound      = errors.New("token not found, expired, or revoked")
	ErrCredentialNotFound = errors.New("credential not found")
	ErrLastCredential     = errors.New("cannot revoke the account's last remaining credential")
)

// User is the dashboard's single account.
type User struct {
	ID          string
	UserHandle  []byte
	DisplayName string
	CreatedAt   time.Time
}

// Token is an API token's metadata -- never its raw secret, which is only
// ever available at creation. See add-api-token-auth/design.md's "Hash the
// token at rest" decision.
type Token struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// CredentialInfo is a WebAuthn credential's metadata for the credential-
// management UI -- never the credential itself (public key, sign count,
// ...), which stays internal to the WebAuthn ceremony. ID is the raw
// credential id; the HTTP layer is what encodes it for the wire (see
// support-multiple-passkeys/design.md's "base64url of the credential ID"
// decision).
type CredentialInfo struct {
	ID        []byte
	Label     string
	CreatedAt time.Time
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
	// PutCredentialLabeled creates a new credential belonging to userID
	// with a user-supplied label -- the authenticated "add a passkey"
	// path, as opposed to PutCredential's anonymous-registration and
	// sign-count-bump uses, neither of which have a label to set.
	PutCredentialLabeled(ctx context.Context, userID string, cred webauthn.Credential, label string) error
	// CredentialsForUser returns every credential belonging to userID.
	CredentialsForUser(ctx context.Context, userID string) ([]webauthn.Credential, error)
	// CredentialInfosForUser returns every credential belonging to userID
	// as its listing metadata (label, creation date), for the credential-
	// management UI.
	CredentialInfosForUser(ctx context.Context, userID string) ([]CredentialInfo, error)
	// RevokeCredential deletes a credential belonging to userID, scoped by
	// its raw id. Returns ErrCredentialNotFound if no such credential
	// exists for that user, or ErrLastCredential if it's the account's
	// only remaining credential -- both checked and applied in a single
	// transaction, so a concurrent revoke of the account's last two
	// credentials can't zero it out (support-multiple-passkeys/design.md's
	// "Concurrent revoke" risk).
	RevokeCredential(ctx context.Context, userID string, credentialID []byte) error

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

	// CreateToken creates a new API token for userID, returning its
	// metadata and its raw secret -- the only time the raw secret is ever
	// available.
	CreateToken(ctx context.Context, userID string) (Token, string, error)
	// TokenUserID returns the userID a raw token belongs to. Returns
	// ErrTokenNotFound if the token doesn't exist, has expired, or has
	// been revoked.
	TokenUserID(ctx context.Context, rawToken string) (string, error)
	// RevokeToken revokes a token by id, scoped to userID. Returns
	// ErrTokenNotFound if no such token exists for that user.
	RevokeToken(ctx context.Context, userID, tokenID string) error
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
