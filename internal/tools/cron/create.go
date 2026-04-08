package cron

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/tools"
)

type createInput struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	Cwd      string `json:"cwd,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
}

// CreateTool creates a new cron job via the Manager.
type CreateTool struct {
	Mgr *Manager
}

func (CreateTool) Name() string { return "cron_create" }

func (CreateTool) Description() string {
	return "Creates a new cron job with a name, schedule expression, and command to run."
}

func (CreateTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Unique name for the cron job",
			},
			"schedule": map[string]any{
				"type":        "string",
				"description": "Cron expression (e.g. '0 * * * *' for every hour)",
			},
			"command": map[string]any{
				"type":        "string",
				"description": "Shell command to execute",
			},
			"cwd": map[string]any{
				"type":        "string",
				"description": "Working directory for command execution",
			},
			"enabled": map[string]any{
				"type":        "boolean",
				"description": "Whether the job is enabled (default: true)",
				"default":     true,
			},
		},
		"required":             []string{"name", "schedule", "command"},
		"additionalProperties": false,
	}
}

func (t CreateTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	var input createInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}

	job := Job{
		Name:     input.Name,
		Schedule: input.Schedule,
		Command:  input.Command,
		Cwd:      input.Cwd,
		Enabled:  enabled,
	}

	if err := t.Mgr.Create(job); err != nil {
		return tools.Result{Content: err.Error(), IsError: true}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Created cron job %q", input.Name)}, nil
}
