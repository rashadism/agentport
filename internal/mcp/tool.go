package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"charm.land/fantasy"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool wraps an MCP tool as a Fantasy AgentTool
type Tool struct {
	manager    *Manager
	serverName string
	tool       *gomcp.Tool
}

func (t *Tool) Info() fantasy.ToolInfo {
	parameters := make(map[string]any)
	required := make([]string, 0)

	if input, ok := t.tool.InputSchema.(map[string]any); ok {
		if props, ok := input["properties"].(map[string]any); ok {
			parameters = props
		}
		if req, ok := input["required"].([]any); ok {
			for _, v := range req {
				if s, ok := v.(string); ok {
					required = append(required, s)
				}
			}
		}
	}

	return fantasy.ToolInfo{
		Name:        fmt.Sprintf("mcp_%s_%s", t.serverName, t.tool.Name),
		Description: t.tool.Description,
		Parameters:  parameters,
		Required:    required,
	}
}

func (t *Tool) Run(ctx context.Context, params fantasy.ToolCall) (fantasy.ToolResponse, error) {
	// Get session with auto-reconnect
	session, err := t.manager.GetSession(ctx, t.serverName)
	if err != nil {
		return fantasy.NewTextErrorResponse(err.Error()), nil
	}

	var args map[string]any
	if err := json.Unmarshal([]byte(params.Input), &args); err != nil {
		return fantasy.NewTextErrorResponse(fmt.Sprintf("error parsing parameters: %v", err)), nil
	}

	result, err := session.CallTool(ctx, &gomcp.CallToolParams{
		Name:      t.tool.Name,
		Arguments: args,
	})
	if err != nil {
		return fantasy.NewTextErrorResponse(err.Error()), nil
	}

	if len(result.Content) == 0 {
		return fantasy.NewTextResponse(""), nil
	}

	var textParts []string
	for _, v := range result.Content {
		switch content := v.(type) {
		case *gomcp.TextContent:
			textParts = append(textParts, content.Text)
		default:
			textParts = append(textParts, fmt.Sprintf("%v", v))
		}
	}

	textContent := strings.Join(textParts, "\n")

	// Apply response transformer if one exists for this tool
	if transformer := GetTransformer(t.tool.Name); transformer != nil {
		var content map[string]any
		if err := json.Unmarshal([]byte(textContent), &content); err == nil {
			if transformed, err := transformer.Transform(content); err == nil && transformed != "" {
				textContent = transformed
			}
		}
	}

	return fantasy.NewTextResponse(textContent), nil
}

func (t *Tool) ProviderOptions() fantasy.ProviderOptions {
	return nil
}

func (t *Tool) SetProviderOptions(_ fantasy.ProviderOptions) {}
