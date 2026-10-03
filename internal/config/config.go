// Package config holds pipeline-analytics' runtime configuration.
package config

import (
	"errors"
	"log/slog"
	"time"
)

// Sentinel validation errors, so a caller can errors.Is against a specific
// missing field instead of matching on message text.
var (
	ErrAddrRequired                 = errors.New("addr must not be empty")
	ErrDBPathRequired               = errors.New("db path must not be empty")
	ErrCallbackURLRequired          = errors.New("callback url must not be empty")
	ErrEncryptionKeyInvalid         = errors.New("encryption key must decode to 32 bytes")
	ErrReconcileIntervalNonPositive = errors.New("reconcile interval must be positive")
	ErrDrainPeriodNegative          = errors.New("drain period must not be negative")
	ErrShutdownTimeoutNonPositive   = errors.New("shutdown timeout must be positive")
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
	// ReconcileInterval is how often reconciliation polling runs against
	// every tracked repo (forge-ingestion/spec.md requires at least
	// hourly).
	ReconcileInterval time.Duration
	// DrainPeriod is how long the server keeps serving after a shutdown
	// signal, with /readyz already answering 503, so a router polling it
	// notices before the listener closes (api.md's /readyz bullet). Zero
	// skips the wait.
	DrainPeriod time.Duration
	// ShutdownTimeout bounds how long in-flight requests get to finish once
	// the drain is over, so a request that never ends can't hold the process
	// open.
	ShutdownTimeout time.Duration
	// LogLevel is the minimum slog level the server logs at.
	LogLevel slog.Level
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

	if c.ReconcileInterval <= 0 {
		return ErrReconcileIntervalNonPositive
	}

	if c.DrainPeriod < 0 {
		return ErrDrainPeriodNegative
	}

	if c.ShutdownTimeout <= 0 {
		return ErrShutdownTimeoutNonPositive
	}

	return nil
}
