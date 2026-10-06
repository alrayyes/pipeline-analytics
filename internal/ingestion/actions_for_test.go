package ingestion_test

import (
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/stretchr/testify/require"
)

func TestActionsFor(t *testing.T) {
	t.Parallel()

	t.Run("a concluded GitHub run can be re-run", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, []ingestion.RunAction{ingestion.ActionRerun}, ingestion.ActionsFor(ingestion.ForgeGitHub, "completed"))
	})

	t.Run("a GitHub run that hasn't concluded can be cancelled", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, []ingestion.RunAction{ingestion.ActionCancel}, ingestion.ActionsFor(ingestion.ForgeGitHub, "in_progress"))
	})

	t.Run("a Forgejo run offers nothing", func(t *testing.T) {
		t.Parallel()

		require.Empty(t, ingestion.ActionsFor(ingestion.ForgeForgejo, "completed"))
	})
}
