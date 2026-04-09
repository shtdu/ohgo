package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/tasks"
	"github.com/shtdu/ohgo/internal/tools"
)

type createInput struct {
	Command         string `json:"command"`
	Description     string `json:"description"`
	Cwd             string `json:"cwd,omitempty"`
	RunInBackground *bool  `json:"run_in_background,omitempty"`
}

// CreateTool creates a new background shell task via the Manager.
type CreateTool struct {
	Mgr *tasks.Manager
}

func (CreateTool) Name() string { return "task_create" }

func (CreateTool) Description() string {
	return "Creates a new background shell task. Returns the task record with ID and status."
}

func (CreateTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "Shell command to execute",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Human-readable description of the task",
			},
			"cwd": map[string]any{
				"type":        "string",
				"description": "Working directory for command execution (default: current directory)",
				"default":     ".",
			},
			"run_in_background": map[string]any{
				"type":        "boolean",
				"description": "Whether to run the task in the background (default: true)",
				"default":     true,
			},
		},
		"required":             []string{"command", "description"},
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

	if input.Command == "" {
		return tools.Result{Content: "invalid arguments: command is required", IsError: true}, nil
	}

	cwd := input.Cwd
	if cwd == "" {
		cwd = "."
	}

	rec, err := t.Mgr.CreateShell(ctx, input.Command, input.Description, cwd)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to create task: %v", err), IsError: true}, nil
	}

	// Re-fetch via Get to obtain a safe copy (background goroutine mutates the
	// original record in place).
	snapshot, ok := t.Mgr.Get(rec.ID)
	if !ok {
		snapshot = rec // fallback to the original
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to marshal result: %v", err), IsError: true}, nil
	}

	return tools.Result{Content: string(data)}, nil
}

var _ tools.Tool = CreateTool{}
