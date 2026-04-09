package team

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/coordinator"
	"github.com/shtdu/ohgo/internal/tools"
)

type deleteInput struct {
	Name string `json:"name"`
}

// DeleteTool removes a team by name via the Coordinator.
type DeleteTool struct {
	Coord *coordinator.Coordinator
}

func (DeleteTool) Name() string { return "team_delete" }

func (DeleteTool) Description() string {
	return "Deletes a named team. Does not stop or remove the agents themselves."
}

func (DeleteTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Name of the team to delete",
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

	if t.Coord == nil {
		return tools.Result{Content: "coordinator not configured", IsError: true}, nil
	}

	var input deleteInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if err := t.Coord.DeleteTeam(input.Name); err != nil {
		return tools.Result{Content: err.Error(), IsError: true}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Deleted team %q", input.Name)}, nil
}
