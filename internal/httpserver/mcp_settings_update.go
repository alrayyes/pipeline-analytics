package httpserver

import (
	"context"
	"fmt"

	"github.com/alrayyes/pipeline-analytics/internal/settings"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// addSettingsUpdateTool registers update_settings, the PATCH /api/settings
// partial update. Its argument is the endpoint's own body: a key present with a
// value sets it, a key present with null clears it back to its default, a key
// left out is untouched. A struct cannot tell null from absent, so the input is
// a map and the tool carries an explicit schema, built from the settings
// package's keys and allowed values rather than copied by hand.
func (h *mcpHandler) addSettingsUpdateTool(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_settings",
		Description: "Change the signed-in account's dashboard settings. Give only the settings to change: a value sets it, null clears it back to its default, and one left out is untouched. Returns the settings after the change. Matches PATCH /api/settings.",
		InputSchema: settingsUpdateSchema(),
	}, h.updateSettings)
}

// settingsUpdateSchema is the input schema: one optional property per setting,
// each a string or null, limited to the setting's allowed values where it has a
// closed set, and nothing else.
func settingsUpdateSchema() *jsonschema.Schema {
	properties := make(map[string]*jsonschema.Schema, len(settings.Keys))

	for _, key := range settings.Keys {
		prop := &jsonschema.Schema{Types: []string{"string", "null"}}

		if allowed := settings.AllowedValues(key); len(allowed) > 0 {
			prop.Enum = make([]any, 0, len(allowed)+1)
			for _, value := range allowed {
				prop.Enum = append(prop.Enum, value)
			}

			prop.Enum = append(prop.Enum, nil)
		}

		if key == settings.KeyPipelinesRepoSelector {
			maxLength := settings.MaxRepoSelectorLength
			prop.MaxLength = &maxLength
		}

		properties[key] = prop
	}

	return &jsonschema.Schema{
		Type:                 "object",
		Properties:           properties,
		AdditionalProperties: &jsonschema.Schema{Not: &jsonschema.Schema{}},
	}
}

func (h *mcpHandler) updateSettings(ctx context.Context, _ *mcp.CallToolRequest, in map[string]*string) (*mcp.CallToolResult, settingsDTO, error) {
	info, ok := authInfoFrom(ctx)
	if !ok || h.settings == nil {
		return nil, settingsDTO{}, fmt.Errorf("%w: settings are not available", errInvalidToolInput)
	}

	got, err := h.settings.Update(ctx, info.UserID, in)
	if err != nil {
		code, message := classifySettingsError(err)

		return nil, settingsDTO{}, mcpToolError{code: code, message: message}
	}

	return nil, toSettingsDTO(got), nil
}
