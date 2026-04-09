package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/tasks"
	"github.com/shtdu/ohgo/internal/tools"
)

type outputInput struct {
	TaskID   string `json:"task_id"`
	MaxBytes *int   `json:"max_bytes,omitempty"`
}

// OutputTool reads the captured output of a background task.
type OutputTool struct {
	Mgr *tasks.Manager
}

func (OutputTool) Name() string { return "task_output" }

func (OutputTool) Description() string {
	return "Reads the captured output of a background task. If output exceeds max_bytes, only the last max_bytes are returned."
}

func (OutputTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"task_id": map[string]any{
				"type":        "string",
				"description": "ID of the task to read output from",
			},
			"max_bytes": map[string]any{
				"type":        "integer",
				"description": "Maximum bytes of output to return (0 or omitted uses default 12000)",
			},
		},
		"required":             []string{"task_id"},
		"additionalProperties": false,
	}
}

func (t OutputTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	var input outputInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.TaskID == "" {
		return tools.Result{Content: "invalid arguments: task_id is required", IsError: true}, nil
	}

	maxBytes := 0
	if input.MaxBytes != nil {
		maxBytes = *input.MaxBytes
	}

	output, err := t.Mgr.ReadOutput(input.TaskID, maxBytes)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to read output: %v", err), IsError: true}, nil
	}

	return tools.Result{Content: output}, nil
}

var _ tools.Tool = OutputTool{}
