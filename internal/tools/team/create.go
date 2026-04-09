// Package team provides tools for managing agent teams via the Coordinator.
package team

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/coordinator"
	"github.com/shtdu/ohgo/internal/tools"
)

type createInput struct {
	Name   string   `json:"name"`
	Agents []string `json:"agents"`
}

// CreateTool creates a named team of agents via the Coordinator.
type CreateTool struct {
	Coord *coordinator.Coordinator
}

func (CreateTool) Name() string { return "team_create" }

func (CreateTool) Description() string {
	return "Creates a named team containing the specified agent names."
}

func (CreateTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Unique name for the team",
			},
			"agents": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "List of agent names to include in the team",
			},
		},
		"required":             []string{"name", "agents"},
		"additionalProperties": false,
	}
}

func (t CreateTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	if t.Coord == nil {
		return tools.Result{Content: "coordinator not configured", IsError: true}, nil
	}

	var input createInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if err := t.Coord.CreateTeam(input.Name, input.Agents); err != nil {
		return tools.Result{Content: err.Error(), IsError: true}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Created team %q with agents %v", input.Name, input.Agents)}, nil
}
