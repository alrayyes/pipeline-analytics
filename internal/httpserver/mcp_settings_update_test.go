package httpserver_test

import (
	"context"
	"strings"
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/settings"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestMCPUpdateSettings(t *testing.T) {
	t.Parallel()

	t.Run("sets a value, and a later read sees it", func(t *testing.T) {
		t.Parallel()

		session := connectMCP(t, newSettingsTestServer(t))

		got := callTool[map[string]any](t, session, "update_settings", map[string]any{"theme": "dark"})
		require.Equal(t, "dark", got["theme"])

		again := callTool[map[string]any](t, session, "get_settings", nil)
		require.Equal(t, "dark", again["theme"])
	})

	t.Run("a key left out is left alone", func(t *testing.T) {
		t.Parallel()

		session := connectMCP(t, newSettingsTestServer(t))
		callTool[map[string]any](t, session, "update_settings", map[string]any{"theme": "dark"})

		got := callTool[map[string]any](t, session, "update_settings", map[string]any{"forgeFilter": "github"})

		require.Equal(t, "dark", got["theme"], "the earlier choice survives")
		require.Equal(t, "github", got["forgeFilter"])
	})

	t.Run("null clears a setting back to its default", func(t *testing.T) {
		t.Parallel()

		session := connectMCP(t, newSettingsTestServer(t))
		callTool[map[string]any](t, session, "update_settings", map[string]any{"telemetryWindow": "30d"})

		got := callTool[map[string]any](t, session, "update_settings", map[string]any{"telemetryWindow": nil})

		require.Equal(t, settings.DefaultTelemetryWindow, got["telemetryWindow"])
	})

	t.Run("a value outside the setting's set is refused", func(t *testing.T) {
		t.Parallel()

		session := connectMCP(t, newSettingsTestServer(t))

		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "update_settings", Arguments: map[string]any{"theme": "neon"}})

		require.NoError(t, err)
		require.True(t, res.IsError)
		require.Equal(t, settings.DefaultTheme, callTool[map[string]any](t, session, "get_settings", nil)["theme"], "nothing was stored")
	})

	t.Run("a key that is not a setting is refused by the schema, before the service", func(t *testing.T) {
		t.Parallel()

		session := connectMCP(t, newSettingsTestServer(t))

		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "update_settings", Arguments: map[string]any{"colour": "red"}})

		require.NoError(t, err)
		require.True(t, res.IsError)
		require.NotContains(t, toolErrorText(t, res), "invalid_body", "the service never saw it: the schema said no")
	})

	t.Run("a repo selector past the length limit is refused by the service", func(t *testing.T) {
		t.Parallel()

		session := connectMCP(t, newSettingsTestServer(t))

		msg := toolFailure(t, session, "update_settings", map[string]any{"pipelinesRepoSelector": strings.Repeat("r", settings.MaxRepoSelectorLength+1)})

		require.NotEmpty(t, msg)
	})

	t.Run("an API token may update them, as the endpoint allows", func(t *testing.T) {
		t.Parallel()

		srv, authStore := newSettingsTestServerWithAuthStore(t)

		userID, err := authStore.Session(context.Background(), srv.sessionCookie.Value)
		require.NoError(t, err)

		_, raw, err := authStore.CreateToken(context.Background(), userID, 0)
		require.NoError(t, err)

		got := callTool[map[string]any](t, connectMCPWithBearer(t, srv, raw), "update_settings", map[string]any{"theme": "light"})

		require.Equal(t, "light", got["theme"])
	})

	t.Run("a build with no settings service says so", func(t *testing.T) {
		t.Parallel()

		msg := toolFailure(t, connectMCP(t, newTestServer(t, nil)), "update_settings", map[string]any{"theme": "dark"})

		require.Contains(t, msg, "settings are not available")
	})
}

// TestUpdateSettingsSchemaMatchesSpec compares the tool's input schema with the
// spec's SettingsUpdate, beyond the property names the generic parity test
// checks: each property's types, its allowed values and the repo selector's
// length limit. The tool's schema is built from the settings package, so this
// is what fails when the spec and the package drift apart.
func TestUpdateSettingsSchemaMatchesSpec(t *testing.T) {
	t.Parallel()

	doc, err := openapi3.NewLoader().LoadFromFile("../../openapi/openapi.yaml")
	require.NoError(t, err)

	spec := doc.Components.Schemas["SettingsUpdate"]
	require.NotNil(t, spec)

	res, err := connectMCP(t, newSettingsTestServer(t)).ListTools(context.Background(), nil)
	require.NoError(t, err)

	var tool map[string]any

	for _, candidate := range res.Tools {
		if candidate.Name != "update_settings" {
			continue
		}

		tool, _ = candidate.InputSchema.(map[string]any)
	}

	require.NotNil(t, tool, "update_settings is listed with an object schema")

	props, _ := tool["properties"].(map[string]any)

	for name, ref := range spec.Value.Properties {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, _ := props[name].(map[string]any)
			require.NotNil(t, got, "the tool declares %s", name)

			require.ElementsMatch(t, []any{"string", "null"}, got["type"], "type")

			if len(ref.Value.Enum) > 0 {
				require.Equal(t, ref.Value.Enum, got["enum"], "allowed values, in the spec's order, null included")
			} else {
				require.Nil(t, got["enum"], "no closed set in the spec, none in the tool")
			}

			if ref.Value.MaxLength != nil {
				require.InDelta(t, float64(*ref.Value.MaxLength), got["maxLength"], 0)
			}
		})
	}
}
