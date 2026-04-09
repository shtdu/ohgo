package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcpmanager "github.com/shtdu/ohgo/internal/mcp"
	"github.com/shtdu/ohgo/internal/tools"
)

// Auth reports the connection status of MCP servers.
type Auth struct {
	Mgr *mcpmanager.Manager
}

func (Auth) Name() string { return "mcp_auth" }

func (Auth) Description() string {
	return "Reports the connection and authentication status of MCP servers."
}

func (Auth) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"server_name": map[string]any{
				"type":        "string",
				"description": "Name of the MCP server to check",
			},
			"action": map[string]any{
				"type":        "string",
				"description": "Action to perform (currently only 'status' is supported)",
				"enum":        []string{"status"},
			},
		},
		"required":             []string{"server_name", "action"},
		"additionalProperties": false,
	}
}

func (t Auth) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
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
		Action     string `json:"action"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	switch input.Action {
	case "status":
		return t.handleStatus(input.ServerName)
	default:
		return tools.Result{
			Content: fmt.Sprintf("unsupported action %q: only 'status' is supported", input.Action),
			IsError: true,
		}, nil
	}
}

func (t Auth) handleStatus(serverName string) (tools.Result, error) {
	if serverName == "" {
		// Return status of all connected servers.
		conns := t.Mgr.List()
		if len(conns) == 0 {
			return tools.Result{Content: "No MCP servers connected."}, nil
		}
		type serverStatus struct {
			Name      string `json:"name"`
			Connected bool   `json:"connected"`
		}
		statuses := make([]serverStatus, 0, len(conns))
		for _, c := range conns {
			statuses = append(statuses, serverStatus{Name: c.Name, Connected: true})
		}
		b, err := json.MarshalIndent(statuses, "", "  ")
		if err != nil {
			return tools.Result{Content: fmt.Sprintf("failed to marshal status: %v", err), IsError: true}, nil
		}
		return tools.Result{Content: string(b)}, nil
	}

	// Check a specific server.
	conn, ok := t.Mgr.Get(serverName)
	if !ok {
		return tools.Result{
			Content: fmt.Sprintf("MCP server %q is not connected.", serverName),
			IsError: true,
		}, nil
	}
	return tools.Result{
		Content: fmt.Sprintf("MCP server %q is connected.", conn.Name),
	}, nil
}
