package httpserver_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// mcpToolEndpoint pairs an MCP tool with the REST operation
// openapi/openapi.yaml already declares for the same data, per
// mcp-endpoint/spec.md's "Tool schemas stay in lockstep with the OpenAPI
// spec".
type mcpToolEndpoint struct {
	tool string
	path string
}

var mcpToolEndpoints = []mcpToolEndpoint{
	{tool: "list_pipelines", path: "/api/pipelines"},
	{tool: "get_pipeline", path: "/api/pipelines/{pipelineId}"},
	{tool: "list_pipeline_steps", path: "/api/pipelines/{pipelineId}/steps"},
	{tool: "list_pipeline_flaky_runs", path: "/api/pipelines/{pipelineId}/flaky-runs"},
	{tool: "get_repo_usage", path: "/api/repos/{repoId}/usage"},
	{tool: "list_unhealthy_steps", path: "/api/steps/unhealthy"},
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
