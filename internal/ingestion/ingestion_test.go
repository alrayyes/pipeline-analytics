package ingestion_test

import (
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/stretchr/testify/require"
)

func TestMaskToken(t *testing.T) {
	t.Parallel()

	t.Run("keeps only the last four characters", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "****cdef", ingestion.MaskToken("ghp_abcdcdef"))
	})

	t.Run("masks entirely when too short to reveal a suffix safely", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, "****", ingestion.MaskToken("abc"))
	})
}
