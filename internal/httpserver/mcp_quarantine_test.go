package httpserver_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// toolFailure calls a tool that is expected to report an error and returns it.
func toolFailure(t *testing.T, session *mcp.ClientSession, tool string, args map[string]any) string {
	t.Helper()

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: tool, Arguments: args})
	require.NoError(t, err)
	require.True(t, res.IsError, "the call should have reported an error")

	return toolErrorText(t, res)
}

type quarantineResult struct {
	Quarantined bool `json:"quarantined"`
	Quarantine  *struct {
		Note          string    `json:"note"`
		QuarantinedAt time.Time `json:"quarantinedAt"`
		ExpiresAt     time.Time `json:"expiresAt"`
	} `json:"quarantine"`
}

func TestMCPQuarantineStep(t *testing.T) {
	t.Parallel()

	t.Run("marks the step for 30 days with a note, and the flaky list reports it", func(t *testing.T) {
		t.Parallel()

		srv, id, _ := flakyServer(t)
		session := connectMCP(t, srv)

		got := callTool[quarantineResult](t, session, "quarantine_step", map[string]any{
			"pipelineId": id, "step": "test", "note": "waits on the shared database",
		})

		require.True(t, got.Quarantined)
		require.NotNil(t, got.Quarantine)
		require.Equal(t, "waits on the shared database", got.Quarantine.Note)
		require.WithinDuration(t, time.Now().Add(30*24*time.Hour), got.Quarantine.ExpiresAt, time.Minute)

		list := listQuarantined(t, srv)
		require.Len(t, list.Steps, 1)
		require.True(t, list.Steps[0].Quarantined, "the endpoint sees the mark the tool set")
	})

	t.Run("a note is optional", func(t *testing.T) {
		t.Parallel()

		srv, id, _ := flakyServer(t)

		got := callTool[quarantineResult](t, connectMCP(t, srv), "quarantine_step", map[string]any{"pipelineId": id, "step": "test"})

		require.True(t, got.Quarantined)
	})

	t.Run("a note over 500 characters is refused as on the endpoint", func(t *testing.T) {
		t.Parallel()

		srv, id, _ := flakyServer(t)

		msg := toolFailure(t, connectMCP(t, srv), "quarantine_step", map[string]any{
			"pipelineId": id, "step": "test", "note": strings.Repeat("x", 501),
		})

		require.Contains(t, msg, "note_too_long")
		require.False(t, listQuarantined(t, srv).Steps[0].Quarantined, "nothing was marked")
	})

	t.Run("an unknown pipeline is not found", func(t *testing.T) {
		t.Parallel()

		srv, _, _ := flakyServer(t)

		msg := toolFailure(t, connectMCP(t, srv), "quarantine_step", map[string]any{"pipelineId": "no-such-pipeline", "step": "test"})

		require.Contains(t, msg, "not_found")
	})

	t.Run("an API token is not a session, so it marks nothing", func(t *testing.T) {
		t.Parallel()

		srv, id, authStore := flakyServer(t)

		userID, err := authStore.Session(context.Background(), srv.sessionCookie.Value)
		require.NoError(t, err)

		_, raw, err := authStore.CreateToken(context.Background(), userID, 0)
		require.NoError(t, err)

		msg := toolFailure(t, connectMCPWithBearer(t, srv, raw), "quarantine_step", map[string]any{"pipelineId": id, "step": "test"})

		require.Contains(t, msg, "unauthenticated")
		require.False(t, listQuarantined(t, srv).Steps[0].Quarantined)
	})
}

func TestMCPUnquarantineStep(t *testing.T) {
	t.Parallel()

	t.Run("clears the mark", func(t *testing.T) {
		t.Parallel()

		srv, id, _ := flakyServer(t)
		session := connectMCP(t, srv)
		callTool[quarantineResult](t, session, "quarantine_step", map[string]any{"pipelineId": id, "step": "test"})
		require.True(t, listQuarantined(t, srv).Steps[0].Quarantined)

		got := callTool[quarantineResult](t, session, "unquarantine_step", map[string]any{"pipelineId": id, "step": "test"})

		require.False(t, got.Quarantined)
		require.False(t, listQuarantined(t, srv).Steps[0].Quarantined)
	})

	t.Run("a step with no mark is fine", func(t *testing.T) {
		t.Parallel()

		srv, id, _ := flakyServer(t)

		got := callTool[quarantineResult](t, connectMCP(t, srv), "unquarantine_step", map[string]any{"pipelineId": id, "step": "test"})

		require.False(t, got.Quarantined)
	})

	t.Run("an API token cannot clear a mark either", func(t *testing.T) {
		t.Parallel()

		srv, id, authStore := flakyServer(t)
		callTool[quarantineResult](t, connectMCP(t, srv), "quarantine_step", map[string]any{"pipelineId": id, "step": "test"})

		userID, err := authStore.Session(context.Background(), srv.sessionCookie.Value)
		require.NoError(t, err)

		_, raw, err := authStore.CreateToken(context.Background(), userID, 0)
		require.NoError(t, err)

		msg := toolFailure(t, connectMCPWithBearer(t, srv, raw), "unquarantine_step", map[string]any{"pipelineId": id, "step": "test"})

		require.Contains(t, msg, "unauthenticated")
		require.True(t, listQuarantined(t, srv).Steps[0].Quarantined, "the mark is still there")
	})
}
