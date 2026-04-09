// Package agent provides tools for spawning subagents via the Coordinator.
package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/coordinator"
	"github.com/shtdu/ohgo/internal/tools"
)

type spawnInput struct {
	AgentName   string `json:"agent_name"`
	Prompt      string `json:"prompt"`
	Description string `json:"description,omitempty"`
}

// SpawnTool spawns a new subagent via the Coordinator.
type SpawnTool struct {
	Coord *coordinator.Coordinator
}

func (SpawnTool) Name() string { return "agent_spawn" }

func (SpawnTool) Description() string {
	return "Spawns a new subagent with the given name, prompt, and optional description."
}

func (SpawnTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"agent_name": map[string]any{
				"type":        "string",
				"description": "Unique name for the spawned agent",
			},
			"prompt": map[string]any{
				"type":        "string",
				"description": "The system prompt or instruction for the agent",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Optional human-readable description of the agent's role",
			},
		},
		"required":             []string{"agent_name", "prompt"},
		"additionalProperties": false,
	}
}

func (t SpawnTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	if t.Coord == nil {
		return tools.Result{Content: "coordinator not configured", IsError: true}, nil
	}

	var input spawnInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	spec := coordinator.AgentSpec{
		Name:        input.AgentName,
		Description: input.Description,
		Prompt:      input.Prompt,
	}

	agentID, err := t.Coord.Spawn(ctx, spec)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("spawn agent %q: %v", input.AgentName, err), IsError: true}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Spawned agent %q with ID %s", input.AgentName, agentID)}, nil
}
