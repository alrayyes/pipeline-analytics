package main

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLoadConfig_HyphenatedFlagsFromEnv guards against a real regression:
// viper's default env-var lookup for a flag named "callback-url" is the
// literal key "PIPELINE_ANALYTICS_CALLBACK-URL" -- not a name a shell,
// Docker, or systemd can set, since environment variable names can't
// contain a hyphen. Every hyphenated flag (callback-url, encryption-key,
// reconcile-interval) was unreachable by environment variable at all until
// this test's fix.
func TestLoadConfig_HyphenatedFlagsFromEnv(t *testing.T) {
	newServeCmd()

	t.Setenv("PIPELINE_ANALYTICS_ADDR", ":9090")
	t.Setenv("PIPELINE_ANALYTICS_DB", "test.db")
	t.Setenv("PIPELINE_ANALYTICS_CALLBACK_URL", "https://example.com")
	t.Setenv("PIPELINE_ANALYTICS_ENCRYPTION_KEY", strings.Repeat("00", 32))
	t.Setenv("PIPELINE_ANALYTICS_RECONCILE_INTERVAL", "30m")

	cfg, err := loadConfig()
	require.NoError(t, err)

	require.Equal(t, "https://example.com", cfg.CallbackURL)
	require.Len(t, cfg.EncryptionKey, 32)
	require.Equal(t, "30m0s", cfg.ReconcileInterval.String())
}

func TestParseLogLevel(t *testing.T) {
	t.Parallel()

	t.Run("parses a valid level case-insensitively", func(t *testing.T) {
		t.Parallel()

		level, err := parseLogLevel("DEBUG")
		require.NoError(t, err)
		require.Equal(t, slog.LevelDebug, level)
	})

	t.Run("rejects an invalid level", func(t *testing.T) {
		t.Parallel()

		_, err := parseLogLevel("verbose")
		require.Error(t, err)
	})
}

// TestLoadConfig_LogLevelDefaultsToInfo guards the --log-level flag's
// default actually reaching Config -- with nothing set, logging should stay
// at Info, unchanged from before this flag existed.
func TestLoadConfig_LogLevelDefaultsToInfo(t *testing.T) {
	newServeCmd()

	t.Setenv("PIPELINE_ANALYTICS_CALLBACK_URL", "https://example.com")
	t.Setenv("PIPELINE_ANALYTICS_ENCRYPTION_KEY", strings.Repeat("00", 32))

	cfg, err := loadConfig()
	require.NoError(t, err)

	require.Equal(t, slog.LevelInfo, cfg.LogLevel)
}

// TestLoadConfig_LogLevelFromEnv guards the flag's actual point: turning
// this up via PIPELINE_ANALYTICS_LOG_LEVEL is what --debug-without-a-
// redeploy means in practice.
func TestLoadConfig_LogLevelFromEnv(t *testing.T) {
	newServeCmd()

	t.Setenv("PIPELINE_ANALYTICS_CALLBACK_URL", "https://example.com")
	t.Setenv("PIPELINE_ANALYTICS_ENCRYPTION_KEY", strings.Repeat("00", 32))
	t.Setenv("PIPELINE_ANALYTICS_LOG_LEVEL", "debug")

	cfg, err := loadConfig()
	require.NoError(t, err)

	require.Equal(t, slog.LevelDebug, cfg.LogLevel)
}

func TestLoadConfig_InvalidLogLevel(t *testing.T) {
	newServeCmd()

	t.Setenv("PIPELINE_ANALYTICS_CALLBACK_URL", "https://example.com")
	t.Setenv("PIPELINE_ANALYTICS_ENCRYPTION_KEY", strings.Repeat("00", 32))
	t.Setenv("PIPELINE_ANALYTICS_LOG_LEVEL", "verbose")

	_, err := loadConfig()
	require.Error(t, err)
}
