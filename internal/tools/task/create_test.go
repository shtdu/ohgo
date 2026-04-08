package task

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tasks"
)

// setup returns a fresh Manager and all six tool instances backed by it.
func setup(t *testing.T) (
	*tasks.Manager,
	CreateTool,
	GetTool,
	ListTool,
	OutputTool,
	StopTool,
	UpdateTool,
) {
	t.Helper()
	mgr := tasks.NewManager()
	return mgr,
		CreateTool{Mgr: mgr},
		GetTool{Mgr: mgr},
		ListTool{Mgr: mgr},
		OutputTool{Mgr: mgr},
		StopTool{Mgr: mgr},
		UpdateTool{Mgr: mgr}
}

// mustJSON marshals v to JSON or fails the test.
func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

// mustCreateTask creates a shell task and returns its ID.
func mustCreateTask(t *testing.T, mgr *tasks.Manager, command, description string) string {
	t.Helper()
	rec, err := mgr.CreateShell(context.Background(), command, description, ".")
	require.NoError(t, err)
	require.NotEmpty(t, rec.ID)
	return rec.ID
}

// ---- CreateTool ----

func TestCreateTool_ValidCommand(t *testing.T) {
	_, create, _, _, _, _, _ := setup(t)
	args := mustJSON(t, map[string]any{
		"command":     "echo hello",
		"description": "say hello",
	})
	result, err := create.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "echo hello")
	assert.Contains(t, result.Content, "say hello")
	assert.Contains(t, result.Content, "running")
}

func TestCreateTool_WithCwd(t *testing.T) {
	_, create, _, _, _, _, _ := setup(t)
	args := mustJSON(t, map[string]any{
		"command":     "pwd",
		"description": "print working directory",
		"cwd":         "/tmp",
	})
	result, err := create.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "/tmp")
}

func TestCreateTool_MissingCommand(t *testing.T) {
	_, create, _, _, _, _, _ := setup(t)
	args := mustJSON(t, map[string]any{
		"description": "no command",
	})
	result, err := create.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "command is required")
}

func TestCreateTool_InvalidJSON(t *testing.T) {
	_, create, _, _, _, _, _ := setup(t)
	result, err := create.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestCreateTool_ContextCancel(t *testing.T) {
	_, create, _, _, _, _, _ := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := mustJSON(t, map[string]any{
		"command":     "echo hello",
		"description": "cancelled",
	})
	_, err := create.Execute(ctx, args)
	assert.Error(t, err)
}
