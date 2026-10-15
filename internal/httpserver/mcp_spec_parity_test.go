package httpserver_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// mcpToolEndpoint pairs an MCP tool with the REST operation
// openapi/openapi.yaml already declares for the same data, per
// mcp-endpoint/spec.md's "Tool schemas stay in lockstep with the OpenAPI
// spec".
type mcpToolEndpoint struct {
	tool      string
	operation string
	path      string
}

var mcpToolEndpoints = []mcpToolEndpoint{
	{tool: "list_pipelines", operation: "listPipelines", path: "/api/pipelines"},
	{tool: "get_pipeline", operation: "getPipeline", path: "/api/pipelines/{pipelineId}"},
	{tool: "list_pipeline_steps", operation: "getPipelineSteps", path: "/api/pipelines/{pipelineId}/steps"},
	{tool: "list_pipeline_flaky_runs", operation: "listFlakyRuns", path: "/api/pipelines/{pipelineId}/flaky-runs"},
	{tool: "get_repo_usage", operation: "getRepoUsage", path: "/api/repos/{repoId}/usage"},
	{tool: "list_unhealthy_steps", operation: "listUnhealthySteps", path: "/api/steps/unhealthy"},
	{tool: "get_failure_insights", operation: "getFailureInsights", path: "/api/insights/failures"},
	{tool: "list_runs", operation: "listRuns", path: "/api/runs"},
	{tool: "get_run_steps", operation: "getRunSteps", path: "/api/runs/{runId}/steps"},
	{tool: "get_job_log", operation: "getJobLog", path: "/api/runs/{runId}/jobs/{jobId}/log"},
	{tool: "list_flaky_steps", operation: "listFlakySteps", path: "/api/steps/flaky"},
	{tool: "list_branches", operation: "listBranches", path: "/api/branches"},
	{tool: "get_version", operation: "getVersion", path: "/api/version"},
	{tool: "list_repos", operation: "listRepos", path: "/api/repos"},
	{tool: "list_repo_identifiers", operation: "listRepoIdentifiers", path: "/api/repos/identifiers"},
}

// agentExcludedExtension is the openapi.yaml extension that takes an
// operation out of MCP and WebMCP parity. Its value is the reason, so the
// exclusion lives on the operation it excuses (rules/mcp.md).
const agentExcludedExtension = "x-agent-excluded"

// pendingMCPTools lists the operations whose tool is still to come under
// #615. It exists so the parity test can land first and go green; each
// follow-up pull request removes the operations it covers, and the last one
// deletes the list. "Nobody asked yet" is not an exclusion, so an operation
// stays here only until its tool ships.
var pendingMCPTools = map[string]bool{
	"registerRepo":               true,
	"discoverRepos":              true,
	"untrackRepo":                true,
	"quarantineStep":             true,
	"unquarantineStep":           true,
	"rerunRun":                   true,
	"cancelRun":                  true,
	"getGitHubRateLimitInsights": true,
	"getSettings":                true,
	"updateSettings":             true,
}

// TestEveryOperationHasAnMCPTool is the other half of
// TestMCPToolSchemasMatchOpenAPI: that one checks the tools that exist, this
// one notices an operation nobody gave a tool. It reads every operation ID
// in openapi.yaml, any method, and wants a tool, an x-agent-excluded reason,
// or (for now) a place on pendingMCPTools.
func TestEveryOperationHasAnMCPTool(t *testing.T) {
	t.Parallel()

	doc, err := openapi3.NewLoader().LoadFromFile("../../openapi/openapi.yaml")
	require.NoError(t, err)

	tools := make(map[string]bool, len(mcpToolEndpoints))
	for _, ep := range mcpToolEndpoints {
		tools[ep.operation] = true
	}

	seen := map[string]bool{}

	for path, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			require.NotEmptyf(t, op.OperationID, "%s %s has no operationId", method, path)
			require.Falsef(t, seen[op.OperationID], "operationId %s is declared twice", op.OperationID)
			seen[op.OperationID] = true

			reason, excluded := op.Extensions[agentExcludedExtension]
			if excluded {
				require.Falsef(t, tools[op.OperationID], "%s is both an MCP tool and %s", op.OperationID, agentExcludedExtension)
				require.Falsef(t, pendingMCPTools[op.OperationID], "%s is both excluded and pending", op.OperationID)

				text, isText := reason.(string)
				require.Truef(t, isText && strings.TrimSpace(text) != "", "%s must be a reason, not %v, on %s", agentExcludedExtension, reason, op.OperationID)

				continue
			}

			if pendingMCPTools[op.OperationID] {
				require.Falsef(t, tools[op.OperationID], "%s has its tool: take it off pendingMCPTools", op.OperationID)

				continue
			}

			require.Truef(t, tools[op.OperationID], "%s %s (%s) has no MCP tool: add one to mcp.go and mcpToolEndpoints, or mark the operation %s with a reason", method, path, op.OperationID, agentExcludedExtension)
		}
	}

	for operation := range pendingMCPTools {
		require.Truef(t, seen[operation], "pendingMCPTools names %s, which openapi.yaml no longer declares", operation)
	}

	for _, ep := range mcpToolEndpoints {
		require.Truef(t, seen[ep.operation], "mcpToolEndpoints names %s, which openapi.yaml no longer declares", ep.operation)
	}
}

// TestMCPToolSchemasMatchOpenAPI catches a tool's input schema drifting
// from the REST parameter list it mirrors: add, rename, or change the
// required-ness of a query/path parameter on one side without updating
// the other, and this fails.
func TestMCPToolSchemasMatchOpenAPI(t *testing.T) {
	t.Parallel()

	doc, err := openapi3.NewLoader().LoadFromFile("../../openapi/openapi.yaml")
	require.NoError(t, err)

	srv := newTestServer(t, nil)
	session := connectMCP(t, srv)

	res, err := session.ListTools(context.Background(), nil)
	require.NoError(t, err)

	tools := make(map[string]map[string]any, len(res.Tools))

	for _, tool := range res.Tools {
		schema, ok := tool.InputSchema.(map[string]any)
		require.Truef(t, ok, "tool %s input schema is %T, want map[string]any", tool.Name, tool.InputSchema)
		tools[tool.Name] = schema
	}

	for _, ep := range mcpToolEndpoints {
		t.Run(ep.tool, func(t *testing.T) {
			t.Parallel()

			pathItem := doc.Paths.Find(ep.path)
			require.NotNilf(t, pathItem, "openapi.yaml has no path %s", ep.path)

			op := pathItem.GetOperation(http.MethodGet)
			require.NotNilf(t, op, "openapi.yaml has no GET %s", ep.path)

			specParams := specQueryAndPathParams(op)

			schema, ok := tools[ep.tool]
			require.Truef(t, ok, "no MCP tool named %s", ep.tool)

			require.Equal(t, specParams, toolInputParams(t, schema),
				"%s's input parameters must match GET %s's declared parameters", ep.tool, ep.path)
		})
	}
}

// specQueryAndPathParams maps each of op's path/query parameter names to
// whether it's required, per openapi.yaml -- what an MCP tool's input
// schema for the same operation should match. Header/cookie parameters
// aren't part of a tool's input (auth is handled outside the tool call),
// so they're excluded.
func specQueryAndPathParams(op *openapi3.Operation) map[string]bool {
	params := make(map[string]bool, len(op.Parameters))

	for _, ref := range op.Parameters {
		p := ref.Value
		if p.In != openapi3.ParameterInPath && p.In != openapi3.ParameterInQuery {
			continue
		}

		params[p.Name] = p.Required
	}

	return params
}

// toolInputParams maps each property name in an MCP tool's JSON input
// schema to whether it's required.
func toolInputParams(t *testing.T, schema map[string]any) map[string]bool {
	t.Helper()

	properties, _ := schema["properties"].(map[string]any)

	required := map[string]bool{}
	if list, ok := schema["required"].([]any); ok {
		for _, r := range list {
			if name, ok := r.(string); ok {
				required[name] = true
			}
		}
	}

	params := make(map[string]bool, len(properties))
	for name := range properties {
		params[name] = required[name]
	}

	return params
}
