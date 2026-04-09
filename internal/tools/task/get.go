package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/tasks"
	"github.com/shtdu/ohgo/internal/tools"
)

type getInput struct {
	TaskID string `json:"task_id"`
}

// GetTool retrieves a task record by ID.
type GetTool struct {
	Mgr *tasks.Manager
}

func (GetTool) Name() string { return "task_get" }

func (GetTool) Description() string {
	return "Retrieves a task record by ID. Returns the full task record including status and metadata."
}

func (GetTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"task_id": map[string]any{
				"type":        "string",
				"description": "ID of the task to retrieve",
			},
		},
		"required":             []string{"task_id"},
		"additionalProperties": false,
	}
}

func (t GetTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	var input getInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.TaskID == "" {
		return tools.Result{Content: "invalid arguments: task_id is required", IsError: true}, nil
	}

	rec, found := t.Mgr.Get(input.TaskID)
	if !found {
		return tools.Result{Content: fmt.Sprintf("task %s not found", input.TaskID), IsError: true}, nil
	}

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to marshal result: %v", err), IsError: true}, nil
	}

	return tools.Result{Content: string(data)}, nil
}

var _ tools.Tool = GetTool{}
