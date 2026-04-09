package team

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/coordinator"
	"github.com/shtdu/ohgo/internal/tools"
)

// Verify both tools satisfy the Tool interface.
var (
	_ tools.Tool = CreateTool{}
	_ tools.Tool = DeleteTool{}
)

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

// ---- CreateTool ----

func TestCreateTool_NilCoord(t *testing.T) {
	tool := CreateTool{}
	args := mustJSON(t, map[string]any{
		"name":   "team-a",
		"agents": []string{"agent1"},
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "coordinator not configured")
}

func TestCreateTool_InvalidJSON(t *testing.T) {
	coord := coordinator.New("/bin/true")
	tool := CreateTool{Coord: coord}

	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestCreateTool_ValidCreate(t *testing.T) {
	coord := coordinator.New("/bin/true")
	tool := CreateTool{Coord: coord}

	args := mustJSON(t, map[string]any{
		"name":   "research",
		"agents": []string{"agent-a", "agent-b"},
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Created team")
	assert.Contains(t, result.Content, "research")
}

func TestCreateTool_DuplicateTeam(t *testing.T) {
	coord := coordinator.New("/bin/true")
	tool := CreateTool{Coord: coord}

	args := mustJSON(t, map[string]any{
		"name":   "dup-team",
		"agents": []string{"a1"},
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	// Create again with same name.
	result, err = tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "already exists")
}

func TestCreateTool_ContextCancel(t *testing.T) {
	coord := coordinator.New("/bin/true")
	tool := CreateTool{Coord: coord}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := mustJSON(t, map[string]any{"name": "x", "agents": []string{"a"}})
	_, err := tool.Execute(ctx, args)
	assert.Error(t, err)
}

func TestCreateTool_Interface(t *testing.T) {
	tool := CreateTool{}
	assert.Equal(t, "team_create", tool.Name())
	assert.NotEmpty(t, tool.Description())
	assert.NotNil(t, tool.InputSchema())
}

// ---- DeleteTool ----

func TestDeleteTool_NilCoord(t *testing.T) {
	tool := DeleteTool{}
	args := mustJSON(t, map[string]any{"name": "team-x"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "coordinator not configured")
}

func TestDeleteTool_InvalidJSON(t *testing.T) {
	coord := coordinator.New("/bin/true")
	tool := DeleteTool{Coord: coord}

	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestDeleteTool_ValidDelete(t *testing.T) {
	coord := coordinator.New("/bin/true")
	require.NoError(t, coord.CreateTeam("remove-me", []string{"a1"}))

	tool := DeleteTool{Coord: coord}
	args := mustJSON(t, map[string]any{"name": "remove-me"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Deleted team")
	assert.Contains(t, result.Content, "remove-me")
}

func TestDeleteTool_NotFound(t *testing.T) {
	coord := coordinator.New("/bin/true")
	tool := DeleteTool{Coord: coord}

	args := mustJSON(t, map[string]any{"name": "ghost"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

func TestDeleteTool_ContextCancel(t *testing.T) {
	coord := coordinator.New("/bin/true")
	tool := DeleteTool{Coord: coord}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := mustJSON(t, map[string]any{"name": "x"})
	_, err := tool.Execute(ctx, args)
	assert.Error(t, err)
}

func TestDeleteTool_Interface(t *testing.T) {
	tool := DeleteTool{}
	assert.Equal(t, "team_delete", tool.Name())
	assert.NotEmpty(t, tool.Description())
	assert.NotNil(t, tool.InputSchema())
}

// ---- Integration: Create then Delete ----

func TestIntegration_CreateThenDelete(t *testing.T) {
	coord := coordinator.New("/bin/true")
	createTool := CreateTool{Coord: coord}
	deleteTool := DeleteTool{Coord: coord}

	// Create a team.
	args := mustJSON(t, map[string]any{
		"name":   "integ-team",
		"agents": []string{"agent-x", "agent-y"},
	})
	result, err := createTool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	// Delete the team.
	delArgs := mustJSON(t, map[string]any{"name": "integ-team"})
	result, err = deleteTool.Execute(context.Background(), delArgs)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	// Deleting again should fail.
	result, err = deleteTool.Execute(context.Background(), delArgs)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}
