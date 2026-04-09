package task

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStopTool_RunningTask(t *testing.T) {
	mgr, _, _, _, _, stop, _ := setup(t)
	id := mustCreateTask(t, mgr, "sleep 30", "stop test")

	args := mustJSON(t, map[string]any{"task_id": id})
	result, err := stop.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "stopped")

	// Give the manager a moment to update status.
	time.Sleep(200 * time.Millisecond)

	// Verify the task is now killed.
	rec, found := mgr.Get(id)
	require.True(t, found)
	assert.Equal(t, "killed", string(rec.Status))
}

func TestStopTool_AlreadyCompleted(t *testing.T) {
	mgr, _, get, _, _, stop, _ := setup(t)
	id := mustCreateTask(t, mgr, "echo quick", "already done")

	// Wait for the task to finish using the robust polling helper.
	waitForTask(t, id, get)

	args := mustJSON(t, map[string]any{"task_id": id})
	result, err := stop.Execute(context.Background(), args)
	require.NoError(t, err)
	// Stop on completed task returns nil error — no-op.
	assert.False(t, result.IsError)
}

func TestStopTool_NotFound(t *testing.T) {
	_, _, _, _, _, stop, _ := setup(t)
	args := mustJSON(t, map[string]any{"task_id": "nonexistent"})
	result, err := stop.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

func TestStopTool_EmptyTaskID(t *testing.T) {
	_, _, _, _, _, stop, _ := setup(t)
	args := mustJSON(t, map[string]any{"task_id": ""})
	result, err := stop.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "task_id is required")
}

func TestStopTool_InvalidJSON(t *testing.T) {
	_, _, _, _, _, stop, _ := setup(t)
	result, err := stop.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestStopTool_ContextCancel(t *testing.T) {
	_, _, _, _, _, stop, _ := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := mustJSON(t, map[string]any{"task_id": "x"})
	_, err := stop.Execute(ctx, args)
	assert.Error(t, err)
}
