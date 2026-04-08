package task

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutputTool_RunningTask(t *testing.T) {
	mgr, _, get, _, output, _, _ := setup(t)
	id := mustCreateTask(t, mgr, "echo hello", "output test")

	// Wait for the task to complete so output is available.
	waitForTask(t, id, get)

	args := mustJSON(t, map[string]any{"task_id": id})
	result, err := output.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "hello")
}

func TestOutputTool_CompletedTask(t *testing.T) {
	mgr, _, get, _, output, _, _ := setup(t)
	id := mustCreateTask(t, mgr, "echo done", "completed output test")

	// Wait for completion.
	waitForTask(t, id, get)

	args := mustJSON(t, map[string]any{"task_id": id})
	result, err := output.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "done")
}

func TestOutputTool_WithMaxBytes(t *testing.T) {
	mgr, _, get, _, output, _, _ := setup(t)
	id := mustCreateTask(t, mgr, "echo hello world", "max bytes test")

	waitForTask(t, id, get)

	args := mustJSON(t, map[string]any{
		"task_id":   id,
		"max_bytes": 5,
	})
	result, err := output.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	// Output should be truncated to last 5 bytes.
	assert.LessOrEqual(t, len(result.Content), 5)
}

func TestOutputTool_NotFound(t *testing.T) {
	_, _, _, _, output, _, _ := setup(t)
	args := mustJSON(t, map[string]any{"task_id": "nonexistent"})
	result, err := output.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

func TestOutputTool_EmptyTaskID(t *testing.T) {
	_, _, _, _, output, _, _ := setup(t)
	args := mustJSON(t, map[string]any{"task_id": ""})
	result, err := output.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "task_id is required")
}

func TestOutputTool_InvalidJSON(t *testing.T) {
	_, _, _, _, output, _, _ := setup(t)
	result, err := output.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestOutputTool_ContextCancel(t *testing.T) {
	_, _, _, _, output, _, _ := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := mustJSON(t, map[string]any{"task_id": "x"})
	_, err := output.Execute(ctx, args)
	assert.Error(t, err)
}

