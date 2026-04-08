package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/tasks"
	"github.com/shtdu/ohgo/internal/tools"
)

type updateInput struct {
	TaskID      string `json:"task_id"`
	Description string `json:"description,omitempty"`
	Progress    *int   `json:"progress,omitempty"`
	StatusNote  string `json:"status_note,omitempty"`
}

// UpdateTool modifies a task's description, progress, or status note.
type UpdateTool struct {
	Mgr *tasks.Manager
}

func (UpdateTool) Name() string { return "task_update" }

func (UpdateTool) Description() string {
	return "Updates a task's description, progress, or status note. Only provided fields are changed."
}

func (UpdateTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"task_id": map[string]any{
				"type":        "string",
				"description": "ID of the task to update",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "New description for the task",
			},
			"progress": map[string]any{
				"type":        "integer",
				"description": "Progress percentage (0-100), omit or set to -1 to leave unchanged",
			},
			"status_note": map[string]any{
				"type":        "string",
				"description": "A status note to attach to the task",
			},
		},
		"required":             []string{"task_id"},
		"additionalProperties": false,
	}
}

func (t UpdateTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	var input updateInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.TaskID == "" {
		return tools.Result{Content: "invalid arguments: task_id is required", IsError: true}, nil
	}

	progress := -1
	if input.Progress != nil {
		progress = *input.Progress
	}

	rec, err := t.Mgr.Update(input.TaskID, input.Description, progress, input.StatusNote)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to update task: %v", err), IsError: true}, nil
	}

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to marshal result: %v", err), IsError: true}, nil
	}

	return tools.Result{Content: string(data)}, nil
}
