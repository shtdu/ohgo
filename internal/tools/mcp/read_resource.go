package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcpmanager "github.com/shtdu/ohgo/internal/mcp"
	"github.com/shtdu/ohgo/internal/tools"
)

// ReadResource reads a resource from a connected MCP server by URI.
type ReadResource struct {
	Mgr *mcpmanager.Manager
}

func (ReadResource) Name() string { return "mcp_read_resource" }

func (ReadResource) Description() string {
	return "Reads the content of a resource from a connected MCP server by URI."
}

func (ReadResource) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"server_name": map[string]any{
				"type":        "string",
				"description": "Name of the connected MCP server",
			},
			"uri": map[string]any{
				"type":        "string",
				"description": "URI of the resource to read",
			},
		},
		"required":             []string{"server_name", "uri"},
		"additionalProperties": false,
	}
}

func (t ReadResource) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	if t.Mgr == nil {
		return tools.Result{Content: "mcp manager not configured", IsError: true}, nil
	}

	var input struct {
		ServerName string `json:"server_name"`
		URI        string `json:"uri"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	result, err := t.Mgr.ReadResource(ctx, input.ServerName, input.URI)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("mcp read resource failed: %v", err), IsError: true}, nil
	}

	// Concatenate text content from all resource contents.
	var text string
	for _, c := range result.Contents {
		if c == nil {
			continue
		}
		if c.Text != "" {
			if text != "" {
				text += "\n"
			}
			text += c.Text
		}
	}
	if text == "" {
		// Fallback: marshal the raw contents as JSON.
		b, _ := json.Marshal(result.Contents)
		text = string(b)
	}
	return tools.Result{Content: text}, nil
}
