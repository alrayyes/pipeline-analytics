package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/alrayyes/pipeline-analytics/internal/ingestion"
	"github.com/alrayyes/pipeline-analytics/internal/metrics"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// errRepoNotFound is a local sentinel for a tool call naming an untracked
// repo -- ingestion.Store's GetRepo doesn't guarantee one at the port
// level, only its sqlite adapter does, and reaching into that adapter's
// error from here would cross the layer the REST handlers already keep
// (repos.go's usageHandler.get treats any GetRepo error as not-found the
// same way).
var errRepoNotFound = errors.New("repo not found")

// mcpHandler adapts the same metrics.Service / ingestion.Store calls the
// REST handlers in this package already make into MCP tools, per
// mcp-endpoint/spec.md's "MCP tools cover the existing read surface" --
// no new domain logic, the same translation pipelines.go's handlers do,
// just to a different wire format.
type mcpHandler struct {
	metrics *metrics.Service
	repos   ingestion.Store
}

// newMCPHandler builds the Streamable HTTP handler for POST /api/mcp,
// mounted inside the same requireSession gate as every other data
// endpoint (server.go). Stateless: true -- each request is handled
// independently, with no server-held session state between calls. This
// server has no need for the standalone SSE stream a stateful session
// would enable (there's nothing it pushes to a client unprompted), and
// statelessness keeps a single-account, always-on server simple.
func newMCPHandler(deps Deps) http.Handler {
	h := &mcpHandler{metrics: deps.Metrics, repos: deps.IngestionStore}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "pipeline-analytics",
		Version: deps.Version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_pipelines",
		Description: "List tracked pipelines and their current health status. Matches GET /api/pipelines.",
	}, h.listPipelines)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_pipeline",
		Description: "Get one pipeline's health status, duration trend, and failure-rate trend. Matches GET /api/pipelines/{pipelineId}.",
	}, h.getPipeline)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_pipeline_steps",
		Description: "List a pipeline's steps ranked by duration contribution, with queue/exec split, failure rate, and flaky flag. Matches GET /api/pipelines/{pipelineId}/steps.",
	}, h.listPipelineSteps)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_pipeline_flaky_runs",
		Description: "List every run in which a named step failed, most recent first. Matches GET /api/pipelines/{pipelineId}/flaky-runs.",
	}, h.listPipelineFlakyRuns)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_repo_usage",
		Description: "Get runner-minutes usage per workflow for a tracked repository. Matches GET /api/repos/{repoId}/usage.",
	}, h.getRepoUsage)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_unhealthy_steps",
		Description: "List pipelines with at least one flaky or failing step, grouped by pipeline. Matches GET /api/steps/unhealthy.",
	}, h.listUnhealthySteps)

	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{Stateless: true})
}

type listPipelinesInput struct {
	RepoID string `json:"repoId,omitempty" jsonschema:"restrict to one tracked repo; omitted returns every repo's pipelines"`
	Forge  string `json:"forge,omitempty" jsonschema:"restrict to one forge; omitted returns every forge"`
	Window string `json:"window,omitempty" jsonschema:"trailing run count the health status is computed over; omitted uses a server-chosen default"`
	Limit  int    `json:"limit,omitempty" jsonschema:"max pipelines to return; omitted returns every matching pipeline"`
	Offset int    `json:"offset,omitempty" jsonschema:"pipelines to skip before the returned page"`
}

func (h *mcpHandler) listPipelines(ctx context.Context, _ *mcp.CallToolRequest, in listPipelinesInput) (*mcp.CallToolResult, pipelineListDTO, error) {
	window := metrics.ParseWindow(in.Window)

	pipelines, hasMore, err := h.metrics.ListPipelines(ctx, window, metrics.PipelineListFilter{
		RepoID: in.RepoID,
		Forge:  in.Forge,
		Limit:  in.Limit,
		Offset: in.Offset,
	})
	if err != nil {
		return nil, pipelineListDTO{}, fmt.Errorf("list pipelines: %w", err)
	}

	dtos := make([]pipelineSummaryDTO, 0, len(pipelines))
	for _, p := range pipelines {
		dtos = append(dtos, toPipelineSummaryDTO(p))
	}

	return nil, pipelineListDTO{Pipelines: dtos, HasMore: hasMore}, nil
}

// pipelineIDWindowInput is get_pipeline's and list_pipeline_steps' shared
// input shape -- both mirror a REST endpoint taking only PipelineId and
// Window (openapi/openapi.yaml).
type pipelineIDWindowInput struct {
	PipelineID string `json:"pipelineId" jsonschema:"the pipeline's opaque id, as returned by list_pipelines"`
	Window     string `json:"window,omitempty" jsonschema:"trailing run count the trend or ranking is computed over; omitted uses a server-chosen default"`
}

func (h *mcpHandler) getPipeline(ctx context.Context, _ *mcp.CallToolRequest, in pipelineIDWindowInput) (*mcp.CallToolResult, pipelineDetailDTO, error) {
	window := metrics.ParseWindow(in.Window)

	detail, err := h.metrics.GetPipeline(ctx, metrics.PipelineID(in.PipelineID), window)
	if err != nil {
		if errors.Is(err, metrics.ErrPipelineNotFound) || errors.Is(err, metrics.ErrInvalidPipelineID) {
			return nil, pipelineDetailDTO{}, fmt.Errorf("%w: %s", metrics.ErrPipelineNotFound, in.PipelineID)
		}

		return nil, pipelineDetailDTO{}, fmt.Errorf("get pipeline: %w", err)
	}

	return nil, pipelineDetailDTO{
		pipelineSummaryDTO: toPipelineSummaryDTO(detail.Pipeline),
		DurationTrend:      toTrendDTO(detail.DurationTrend),
		FailureRateTrend:   toTrendDTO(detail.FailureRateTrend),
	}, nil
}

type pipelineStepsResultDTO struct {
	Steps []stepDTO `json:"steps"`
}

// listPipelineSteps mirrors pipelinesHandler.steps: an unknown pipelineId
// isn't distinguished from one with no recorded steps yet -- both report
// an empty list. Only a malformed id is a tool error.
func (h *mcpHandler) listPipelineSteps(ctx context.Context, _ *mcp.CallToolRequest, in pipelineIDWindowInput) (*mcp.CallToolResult, pipelineStepsResultDTO, error) {
	window := metrics.ParseWindow(in.Window)

	steps, err := h.metrics.GetPipelineSteps(ctx, metrics.PipelineID(in.PipelineID), window)
	if err != nil {
		if errors.Is(err, metrics.ErrInvalidPipelineID) {
			return nil, pipelineStepsResultDTO{}, fmt.Errorf("%w: %s", metrics.ErrInvalidPipelineID, in.PipelineID)
		}

		return nil, pipelineStepsResultDTO{}, fmt.Errorf("get pipeline steps: %w", err)
	}

	dtos := make([]stepDTO, 0, len(steps))
	for _, s := range steps {
		dtos = append(dtos, toStepDTO(s))
	}

	return nil, pipelineStepsResultDTO{Steps: dtos}, nil
}

type flakyRunsInput struct {
	PipelineID string `json:"pipelineId" jsonschema:"the pipeline's opaque id, as returned by list_pipelines"`
	Step       string `json:"step" jsonschema:"a step's name, exactly as list_pipeline_steps reports it"`
	Window     string `json:"window,omitempty" jsonschema:"trailing run count to search within; omitted uses a server-chosen default"`
}

type flakyRunsResultDTO struct {
	Runs []flakyRunDTO `json:"runs"`
}

func (h *mcpHandler) listPipelineFlakyRuns(ctx context.Context, _ *mcp.CallToolRequest, in flakyRunsInput) (*mcp.CallToolResult, flakyRunsResultDTO, error) {
	window := metrics.ParseWindow(in.Window)

	runs, err := h.metrics.ListFlakyRuns(ctx, metrics.PipelineID(in.PipelineID), in.Step, window)
	if err != nil {
		if errors.Is(err, metrics.ErrInvalidPipelineID) {
			return nil, flakyRunsResultDTO{}, fmt.Errorf("%w: %s", metrics.ErrInvalidPipelineID, in.PipelineID)
		}

		return nil, flakyRunsResultDTO{}, fmt.Errorf("list flaky runs: %w", err)
	}

	dtos := make([]flakyRunDTO, 0, len(runs))
	for _, run := range runs {
		dtos = append(dtos, toFlakyRunDTO(run))
	}

	return nil, flakyRunsResultDTO{Runs: dtos}, nil
}

type repoUsageInput struct {
	RepoID string `json:"repoId" jsonschema:"the tracked repository's id, as returned by list_pipelines' repoId field"`
	Window string `json:"window,omitempty" jsonschema:"trailing run count usage is summed over; omitted uses a server-chosen default"`
}

type usageResultDTO struct {
	Usage []usageEntryDTO `json:"usage"`
}

func (h *mcpHandler) getRepoUsage(ctx context.Context, _ *mcp.CallToolRequest, in repoUsageInput) (*mcp.CallToolResult, usageResultDTO, error) {
	if _, err := h.repos.GetRepo(ctx, in.RepoID); err != nil {
		return nil, usageResultDTO{}, fmt.Errorf("%w: %s", errRepoNotFound, in.RepoID)
	}

	window := metrics.ParseWindow(in.Window)

	usage, err := h.metrics.GetRepoUsage(ctx, in.RepoID, window)
	if err != nil {
		return nil, usageResultDTO{}, fmt.Errorf("get repo usage: %w", err)
	}

	dtos := make([]usageEntryDTO, 0, len(usage))
	for _, u := range usage {
		dtos = append(dtos, usageEntryDTO{Workflow: u.Workflow, RunnerMinutes: u.RunnerMinutes})
	}

	return nil, usageResultDTO{Usage: dtos}, nil
}

type unhealthyStepsInput struct {
	Window string `json:"window,omitempty" jsonschema:"trailing run count each pipeline's steps are aggregated over; omitted uses a server-chosen default"`
	Limit  int    `json:"limit,omitempty" jsonschema:"max pipeline groups to return; omitted returns every matching group"`
	Offset int    `json:"offset,omitempty" jsonschema:"pipeline groups to skip before the returned page"`
}

func (h *mcpHandler) listUnhealthySteps(ctx context.Context, _ *mcp.CallToolRequest, in unhealthyStepsInput) (*mcp.CallToolResult, pipelineStepsGroupListDTO, error) {
	window := metrics.ParseWindow(in.Window)

	groups, hasMore, err := h.metrics.ListUnhealthySteps(ctx, window, in.Limit, in.Offset)
	if err != nil {
		return nil, pipelineStepsGroupListDTO{}, fmt.Errorf("list unhealthy steps: %w", err)
	}

	dtos := make([]pipelineStepsGroupDTO, 0, len(groups))

	for _, g := range groups {
		steps := make([]stepDTO, 0, len(g.Steps))
		for _, s := range g.Steps {
			steps = append(steps, toStepDTO(s))
		}

		dtos = append(dtos, pipelineStepsGroupDTO{
			PipelineID:   string(g.PipelineID),
			PipelineName: g.PipelineName,
			RepoID:       g.RepoID,
			Steps:        steps,
		})
	}

	return nil, pipelineStepsGroupListDTO{Groups: dtos, HasMore: hasMore}, nil
}
