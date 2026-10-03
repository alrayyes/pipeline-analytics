package config_test

import (
	"testing"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/config"
	"github.com/stretchr/testify/require"
)

func validConfig() config.Config {
	return config.Config{
		Addr:              ":8080",
		DBPath:            "pipeline-analytics.db",
		CallbackURL:       "https://example.com",
		EncryptionKey:     make([]byte, 32),
		ReconcileInterval: time.Hour,
		ShutdownTimeout:   5 * time.Second,
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, validConfig().Validate())
	})

	t.Run("a zero drain period is allowed, a negative one is not", func(t *testing.T) {
		t.Parallel()

		cfg := validConfig()
		cfg.DrainPeriod = 0
		require.NoError(t, cfg.Validate())

		cfg.DrainPeriod = -time.Second
		require.ErrorIs(t, cfg.Validate(), config.ErrDrainPeriodNegative)
	})

	t.Run("the shutdown timeout must be positive, or shutdown can wait forever", func(t *testing.T) {
		t.Parallel()

		cfg := validConfig()
		cfg.ShutdownTimeout = 0
		require.ErrorIs(t, cfg.Validate(), config.ErrShutdownTimeoutNonPositive)
	})

	t.Run("missing addr", func(t *testing.T) {
		t.Parallel()

		cfg := validConfig()
		cfg.Addr = ""
		require.ErrorIs(t, cfg.Validate(), config.ErrAddrRequired)
	})

	t.Run("missing db path", func(t *testing.T) {
		t.Parallel()

		cfg := validConfig()
		cfg.DBPath = ""
		require.ErrorIs(t, cfg.Validate(), config.ErrDBPathRequired)
	})

	t.Run("missing callback url", func(t *testing.T) {
		t.Parallel()

		cfg := validConfig()
		cfg.CallbackURL = ""
		require.ErrorIs(t, cfg.Validate(), config.ErrCallbackURLRequired)
	})

	t.Run("encryption key wrong size", func(t *testing.T) {
		t.Parallel()

		cfg := validConfig()
		cfg.EncryptionKey = make([]byte, 16)
		require.ErrorIs(t, cfg.Validate(), config.ErrEncryptionKeyInvalid)
	})

	t.Run("non-positive reconcile interval", func(t *testing.T) {
		t.Parallel()

		cfg := validConfig()
		cfg.ReconcileInterval = 0
		require.ErrorIs(t, cfg.Validate(), config.ErrReconcileIntervalNonPositive)
	})
}
