// Package mcp provides tools that interact with MCP (Model Context Protocol) servers.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpmanager "github.com/shtdu/ohgo/internal/mcp"
	"github.com/shtdu/ohgo/internal/tools"
)

// CallTool invokes a tool on a connected MCP server.
type CallTool struct {
	Mgr *mcpmanager.Manager
}

func (CallTool) Name() string { return "mcp_call_tool" }

func (CallTool) Description() string {
	return "Calls a tool on a connected MCP server and returns the text result."
}

func (CallTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"server_name": map[string]any{
				"type":        "string",
				"description": "Name of the connected MCP server",
			},
			"tool_name": map[string]any{
				"type":        "string",
				"description": "Name of the tool to call on the server",
			},
			"arguments": map[string]any{
				"type":        "object",
				"description": "Arguments to pass to the tool",
			},
		},
		"required":             []string{"server_name", "tool_name", "arguments"},
		"additionalProperties": false,
	}
}

func (t CallTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	if t.Mgr == nil {
		return tools.Result{Content: "mcp manager not configured", IsError: true}, nil
	}

	var input struct {
		ServerName string         `json:"server_name"`
		ToolName   string         `json:"tool_name"`
		Arguments  map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	result, err := t.Mgr.CallTool(ctx, input.ServerName, input.ToolName, input.Arguments)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("mcp call tool failed: %v", err), IsError: true}, nil
	}

	text := extractTextContent(result.Content)
	if result.IsError {
		return tools.Result{Content: text, IsError: true}, nil
	}
	return tools.Result{Content: text}, nil
}

// extractTextContent concatenates all TextContent items from a Content list.
func extractTextContent(contents []mcpsdk.Content) string {
	var text string
	for _, c := range contents {
		if tc, ok := c.(*mcpsdk.TextContent); ok && tc != nil {
			if text != "" {
				text += "\n"
			}
			text += tc.Text
		}
	}
	if text == "" {
		// Fallback: marshal the raw contents as JSON.
		b, _ := json.Marshal(contents)
		return string(b)
	}
	return text
}
