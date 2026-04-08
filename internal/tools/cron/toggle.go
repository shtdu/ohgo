package cron

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/tools"
)

type toggleInput struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// ToggleTool enables or disables a cron job via the Manager.
type ToggleTool struct {
	Mgr *Manager
}

func (ToggleTool) Name() string { return "cron_toggle" }

func (ToggleTool) Description() string {
	return "Enables or disables an existing cron job by name."
}

func (ToggleTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Name of the cron job to toggle",
			},
			"enabled": map[string]any{
				"type":        "boolean",
				"description": "True to enable, false to disable",
			},
		},
		"required":             []string{"name", "enabled"},
		"additionalProperties": false,
	}
}

func (t ToggleTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	var input toggleInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if !t.Mgr.Toggle(input.Name, input.Enabled) {
		return tools.Result{Content: fmt.Sprintf("cron job %q not found", input.Name), IsError: true}, nil
	}

	status := "disabled"
	if input.Enabled {
		status = "enabled"
	}
	return tools.Result{Content: fmt.Sprintf("Cron job %q %s", input.Name, status)}, nil
}
