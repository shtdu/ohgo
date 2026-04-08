package task

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTool_Existing(t *testing.T) {
	mgr, _, get, _, _, _, _ := setup(t)
	id := mustCreateTask(t, mgr, "echo hello", "test get task")

	args := mustJSON(t, map[string]any{"task_id": id})
	result, err := get.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, id)
	assert.Contains(t, result.Content, "test get task")
}

func TestGetTool_NotFound(t *testing.T) {
	_, _, get, _, _, _, _ := setup(t)
	args := mustJSON(t, map[string]any{"task_id": "nonexistent"})
	result, err := get.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

func TestGetTool_EmptyID(t *testing.T) {
	_, _, get, _, _, _, _ := setup(t)
	args := mustJSON(t, map[string]any{"task_id": ""})
	result, err := get.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "task_id is required")
}

func TestGetTool_InvalidJSON(t *testing.T) {
	_, _, get, _, _, _, _ := setup(t)
	result, err := get.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestGetTool_ContextCancel(t *testing.T) {
	_, _, get, _, _, _, _ := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := mustJSON(t, map[string]any{"task_id": "x"})
	_, err := get.Execute(ctx, args)
	assert.Error(t, err)
}

// waitForTask polls the Get tool until the task reaches a terminal state.
func waitForTask(t *testing.T, taskID string, get GetTool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		result, _ := get.Execute(context.Background(), mustJSON(t, map[string]any{"task_id": taskID}))
		if !result.IsError && (strings.Contains(result.Content, "completed") || strings.Contains(result.Content, "failed")) {
			// Give a small grace period for output file to be fully flushed.
			time.Sleep(100 * time.Millisecond)
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("timed out waiting for task to finish")
}
