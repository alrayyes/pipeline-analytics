package httpserver

import (
	"context"
	"time"

	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// addQuarantineTools registers the two tools that mark a flaky step as known
// and clear the mark. Like the endpoints they are session-only: a caller
// holding only an API token is refused before anything is written
// (requireToolSession). Neither touches the forge.
func (h *mcpHandler) addQuarantineTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "quarantine_step",
		Description: "Mark a flaky step as known for 30 days, with an optional note, so it stops raising the flaky_step health signal. Marking a step that is already quarantined renews it. Needs a session, not an API token. Matches PUT /api/pipelines/{pipelineId}/steps/{step}/quarantine.",
	}, h.quarantineStep)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "unquarantine_step",
		Description: "Clear a step's quarantine so it raises the flaky_step health signal again. A step with no mark is fine. Needs a session, not an API token. Matches DELETE /api/pipelines/{pipelineId}/steps/{step}/quarantine.",
	}, h.unquarantineStep)
}

type stepInput struct {
	PipelineID string `json:"pipelineId" jsonschema:"the pipeline's id, as list_pipelines reports it"`
	Step       string `json:"step" jsonschema:"the step's name, as list_flaky_steps reports it"`
}

type quarantineStepInput struct {
	stepInput

	Note string `json:"note,omitempty" jsonschema:"why the step is known to be flaky; at most 500 characters"`
}

// checkStepInput refuses an empty pipeline or step. The endpoints' path cannot
// match an empty segment, so a tool must not write a mark for one either.
func checkStepInput(in stepInput) error {
	if in.PipelineID == "" || in.Step == "" {
		return mcpToolError{code: "invalid_input", message: "pipelineId and step are required"}
	}

	return nil
}

func (h *mcpHandler) quarantineStep(ctx context.Context, _ *mcp.CallToolRequest, in quarantineStepInput) (*mcp.CallToolResult, stepQuarantineDTO, error) {
	if err := requireToolSession(ctx, "quarantine a step"); err != nil {
		return nil, stepQuarantineDTO{}, err
	}

	if err := checkStepInput(in.stepInput); err != nil {
		return nil, stepQuarantineDTO{}, err
	}

	q, err := h.metrics.QuarantineStep(ctx, metrics.PipelineID(in.PipelineID), in.Step, in.Note, time.Now())
	if err != nil {
		_, code, message := classifyQuarantineError(err)

		return nil, stepQuarantineDTO{}, mcpToolError{code: code, message: message}
	}

	return nil, stepQuarantineDTO{Quarantined: true, Quarantine: toQuarantineDTO(&q)}, nil
}

func (h *mcpHandler) unquarantineStep(ctx context.Context, _ *mcp.CallToolRequest, in stepInput) (*mcp.CallToolResult, stepQuarantineDTO, error) {
	if err := requireToolSession(ctx, "un-quarantine a step"); err != nil {
		return nil, stepQuarantineDTO{}, err
	}

	if err := checkStepInput(in); err != nil {
		return nil, stepQuarantineDTO{}, err
	}

	if err := h.metrics.UnquarantineStep(ctx, metrics.PipelineID(in.PipelineID), in.Step); err != nil {
		_, code, message := classifyQuarantineError(err)

		return nil, stepQuarantineDTO{}, mcpToolError{code: code, message: message}
	}

	return nil, stepQuarantineDTO{}, nil
}
