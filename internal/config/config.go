// Package config holds pipeline-analytics' runtime configuration.
package config

import "errors"

// Sentinel validation errors, so a caller can errors.Is against a specific
// missing field instead of matching on message text.
var (
	ErrAddrRequired         = errors.New("addr must not be empty")
	ErrDBPathRequired       = errors.New("db path must not be empty")
	ErrCallbackURLRequired  = errors.New("callback url must not be empty")
	ErrEncryptionKeyInvalid = errors.New("encryption key must decode to 32 bytes")
)

// Config is the server's runtime configuration, sourced from flags, the
// environment, and (optionally) a config file, in that precedence.
type Config struct {
	Addr   string
	DBPath string
	// CallbackURL is this server's own public base URL, used to build the
	// webhook callback URL registered on each tracked repo.
	CallbackURL string
	// EncryptionKey is the 32-byte AES-256 key repo tokens are encrypted
	// under at rest.
	EncryptionKey []byte
}

// Validate reports the first invalid or missing required field, if any.
func (c Config) Validate() error {
	if c.Addr == "" {
		return ErrAddrRequired
	}

	if c.DBPath == "" {
		return ErrDBPathRequired
	}

	if c.CallbackURL == "" {
		return ErrCallbackURLRequired
	}

	if len(c.EncryptionKey) != 32 {
		return ErrEncryptionKeyInvalid
	}

	return nil
}
