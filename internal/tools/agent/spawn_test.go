package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/coordinator"
	"github.com/shtdu/ohgo/internal/tools"
)

// Verify SpawnTool satisfies the Tool interface.
var _ tools.Tool = SpawnTool{}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func TestSpawnTool_NilCoord(t *testing.T) {
	tool := SpawnTool{}
	args := mustJSON(t, map[string]any{
		"agent_name": "test-agent",
		"prompt":     "do something",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "coordinator not configured")
}

func TestSpawnTool_InvalidJSON(t *testing.T) {
	coord := coordinator.New("/bin/echo")
	tool := SpawnTool{Coord: coord}

	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestSpawnTool_ValidSpawn(t *testing.T) {
	coord := coordinator.New("/bin/echo")
	tool := SpawnTool{Coord: coord}

	args := mustJSON(t, map[string]any{
		"agent_name": "worker",
		"prompt":     "process data",
		"description": "data processing agent",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Spawned agent")
	assert.Contains(t, result.Content, "worker")
}

func TestSpawnTool_ContextCancel(t *testing.T) {
	coord := coordinator.New("/bin/echo")
	tool := SpawnTool{Coord: coord}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := mustJSON(t, map[string]any{
		"agent_name": "cancelled",
		"prompt":     "noop",
	})
	_, err := tool.Execute(ctx, args)
	assert.Error(t, err)
}

func TestSpawnTool_Interface(t *testing.T) {
	tool := SpawnTool{}
	assert.Equal(t, "agent_spawn", tool.Name())
	assert.NotEmpty(t, tool.Description())
	assert.NotNil(t, tool.InputSchema())
}
