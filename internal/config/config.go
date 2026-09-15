// Package config holds pipeline-analytics' runtime configuration.
package config

import "errors"

// Sentinel validation errors, so a caller can errors.Is against a specific
// missing field instead of matching on message text.
var (
	ErrAddrRequired   = errors.New("addr must not be empty")
	ErrDBPathRequired = errors.New("db path must not be empty")
)

// Config is the server's runtime configuration, sourced from flags, the
// environment, and (optionally) a config file, in that precedence.
type Config struct {
	Addr   string
	DBPath string
}

// Validate reports the first missing required field, if any.
func (c Config) Validate() error {
	if c.Addr == "" {
		return ErrAddrRequired
	}

	if c.DBPath == "" {
		return ErrDBPathRequired
	}

	return nil
}
