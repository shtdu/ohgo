package task

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/tasks"
	"github.com/shtdu/ohgo/internal/tools"
)

type stopInput struct {
	TaskID string `json:"task_id"`
}

// StopTool terminates a running background task.
type StopTool struct {
	Mgr *tasks.Manager
}

func (StopTool) Name() string { return "task_stop" }

func (StopTool) Description() string {
	return "Stops a running background task by sending SIGTERM (then SIGKILL after grace period)."
}

func (StopTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"task_id": map[string]any{
				"type":        "string",
				"description": "ID of the task to stop",
			},
		},
		"required":             []string{"task_id"},
		"additionalProperties": false,
	}
}

func (t StopTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	var input stopInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.TaskID == "" {
		return tools.Result{Content: "invalid arguments: task_id is required", IsError: true}, nil
	}

	err := t.Mgr.Stop(ctx, input.TaskID)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to stop task: %v", err), IsError: true}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Task %s stopped.", input.TaskID)}, nil
}
