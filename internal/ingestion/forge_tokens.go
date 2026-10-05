package ingestion

import (
	"context"
	"errors"
)

// ErrSavedTokenNotFound is returned when no saved token matches: for a forge
// and instance, or for an id.
var ErrSavedTokenNotFound = errors.New("saved token not found")

// SavedToken is a forge token kept for reuse across registrations. Only its
// masked form is ever exposed; the token itself never leaves ForgeToken.
type SavedToken struct {
	ID                 string
	Forge              Forge
	ForgejoInstanceURL string
	TokenMasked        string
}

// ForgeTokenStore keeps at most one token per forge and, for Forgejo,
// instance URL (openspec/specs/forge-ingestion, "A forge token can be saved
// and reused"). Tokens are encrypted at rest.
type ForgeTokenStore interface {
	// SaveForgeToken stores token for the forge and instance, replacing any
	// already saved there. instanceURL is ignored for GitHub.
	SaveForgeToken(ctx context.Context, forge Forge, instanceURL, token string) (SavedToken, error)
	// ListForgeTokens returns every saved token, masked.
	ListForgeTokens(ctx context.Context) ([]SavedToken, error)
	// DeleteForgeToken removes one, or returns ErrSavedTokenNotFound.
	DeleteForgeToken(ctx context.Context, id string) error
	// ForgeToken returns the decrypted token for the forge and instance, or
	// ErrSavedTokenNotFound. For internal use only, never to a client.
	ForgeToken(ctx context.Context, forge Forge, instanceURL string) (string, error)
}
