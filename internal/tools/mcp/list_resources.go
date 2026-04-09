package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcpmanager "github.com/shtdu/ohgo/internal/mcp"
	"github.com/shtdu/ohgo/internal/tools"
)

// ListResources lists all resources available on a connected MCP server.
type ListResources struct {
	Mgr *mcpmanager.Manager
}

func (ListResources) Name() string { return "mcp_list_resources" }

func (ListResources) Description() string {
	return "Lists all resources available on a connected MCP server."
}

func (ListResources) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"server_name": map[string]any{
				"type":        "string",
				"description": "Name of the connected MCP server",
			},
		},
		"required":             []string{"server_name"},
		"additionalProperties": false,
	}
}

func (t ListResources) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
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
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	result, err := t.Mgr.ListResources(ctx, input.ServerName)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("mcp list resources failed: %v", err), IsError: true}, nil
	}

	// Build a simplified list of resource URIs and names for readability.
	type resourceInfo struct {
		URI  string `json:"uri"`
		Name string `json:"name"`
	}
	infos := make([]resourceInfo, 0, len(result.Resources))
	for _, r := range result.Resources {
		if r == nil {
			continue
		}
		infos = append(infos, resourceInfo{URI: r.URI, Name: r.Name})
	}
	b, err := json.MarshalIndent(infos, "", "  ")
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to marshal resources: %v", err), IsError: true}, nil
	}
	return tools.Result{Content: string(b)}, nil
}
