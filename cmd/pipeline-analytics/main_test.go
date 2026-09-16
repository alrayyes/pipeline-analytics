package main

import (
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
