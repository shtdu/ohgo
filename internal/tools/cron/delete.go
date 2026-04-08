package cron

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/tools"
)

type deleteInput struct {
	Name string `json:"name"`
}

// DeleteTool removes a cron job by name via the Manager.
type DeleteTool struct {
	Mgr *Manager
}

func (DeleteTool) Name() string { return "cron_delete" }

func (DeleteTool) Description() string {
	return "Deletes an existing cron job by name."
}

func (DeleteTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Name of the cron job to delete",
			},
		},
		"required":             []string{"name"},
		"additionalProperties": false,
	}
}

func (t DeleteTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	var input deleteInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if !t.Mgr.Delete(input.Name) {
		return tools.Result{Content: fmt.Sprintf("cron job %q not found", input.Name), IsError: true}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Deleted cron job %q", input.Name)}, nil
}
