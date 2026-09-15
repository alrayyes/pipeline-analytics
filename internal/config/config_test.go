package config_test

import (
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/config"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{Addr: ":8080", DBPath: "pipeline-analytics.db"}
		require.NoError(t, cfg.Validate())
	})

	t.Run("missing addr", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{DBPath: "pipeline-analytics.db"}
		require.ErrorIs(t, cfg.Validate(), config.ErrAddrRequired)
	})

	t.Run("missing db path", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{Addr: ":8080"}
		require.ErrorIs(t, cfg.Validate(), config.ErrDBPathRequired)
	})
}
