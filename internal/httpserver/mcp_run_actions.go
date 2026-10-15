package httpserver

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// errRunActionsUnavailable is returned by the run-action tools in a build with
// no run-action service, which leaves the REST routes unmounted.
var errRunActionsUnavailable = errors.New("run actions are not available")

// addRunActionTools registers the two tools that write to a forge. They go
// through the same RunActionService as POST /api/runs/{runId}/rerun and
// /cancel and apply the same rule: a session is required, an API token is
// refused (run_actions.go), so an agent holding only a token cannot re-run or
// cancel anything. A caller with a session, such as the in-browser agent, can.
func (h *mcpHandler) addRunActionTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rerun_run",
		Description: "Ask the forge to re-run a concluded run: only the failed jobs when the run failed, the whole run otherwise. Needs a session, not an API token, and a saved token with Actions write. Matches POST /api/runs/{runId}/rerun.",
	}, h.act("rerun", func(ctx context.Context, runID string) error { return h.runActions.Rerun(ctx, runID) }))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "cancel_run",
		Description: "Ask the forge to cancel a run that has not concluded. Needs a session, not an API token, and a saved token with Actions write. Matches POST /api/runs/{runId}/cancel.",
	}, h.act("cancel", func(ctx context.Context, runID string) error { return h.runActions.Cancel(ctx, runID) }))
}

type runActionInput struct {
	RunID string `json:"runId" jsonschema:"the run's id, as list_runs reports it"`
}

type runActionResult struct {
	RunID    string `json:"runId"`
	Action   string `json:"action"`
	Accepted bool   `json:"accepted"`
}

// act builds a tool handler around one action. The forge acts on its own
// time, so accepted means the forge took the request, not that the run has
// changed yet.
func (h *mcpHandler) act(action string, do func(context.Context, string) error) func(context.Context, *mcp.CallToolRequest, runActionInput) (*mcp.CallToolResult, runActionResult, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in runActionInput) (*mcp.CallToolResult, runActionResult, error) {
		if h.runActions == nil {
			return nil, runActionResult{}, errRunActionsUnavailable
		}

		if err := requireToolSession(ctx, "act on a run"); err != nil {
			return nil, runActionResult{}, err
		}

		if err := do(ctx, in.RunID); err != nil {
			failure := classifyRunActionError(err)

			return nil, runActionResult{}, mcpToolError{code: failure.code, message: failure.message}
		}

		return nil, runActionResult{RunID: in.RunID, Action: action, Accepted: true}, nil
	}
}
