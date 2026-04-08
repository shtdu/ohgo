package task

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tasks"
)

func TestListTool_Empty(t *testing.T) {
	_, _, _, list, _, _, _ := setup(t)
	result, err := list.Execute(context.Background(), json.RawMessage(`{}`))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "No tasks found")
}

func TestListTool_MultipleTasks(t *testing.T) {
	mgr, _, _, list, _, _, _ := setup(t)
	mustCreateTask(t, mgr, "echo alpha", "first task")
	mustCreateTask(t, mgr, "echo beta", "second task")

	result, err := list.Execute(context.Background(), json.RawMessage(`{}`))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "first task")
	assert.Contains(t, result.Content, "second task")
}

func TestListTool_StatusFilter(t *testing.T) {
	mgr, _, _, list, _, _, _ := setup(t)
	mustCreateTask(t, mgr, "echo hello", "running task")

	// Filter for completed tasks — should find none since they are running.
	args := mustJSON(t, map[string]any{"status": string(tasks.StatusCompleted)})
	result, err := list.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "No tasks found")
	assert.Contains(t, result.Content, "completed")
}

func TestListTool_StatusFilterRunning(t *testing.T) {
	mgr, _, _, list, _, _, _ := setup(t)
	mustCreateTask(t, mgr, "sleep 10", "long running")

	args := mustJSON(t, map[string]any{"status": string(tasks.StatusRunning)})
	result, err := list.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "long running")
}

func TestListTool_TableHeader(t *testing.T) {
	mgr, _, _, list, _, _, _ := setup(t)
	mustCreateTask(t, mgr, "echo test", "table test")

	result, err := list.Execute(context.Background(), json.RawMessage(`{}`))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "ID")
	assert.Contains(t, result.Content, "Type")
	assert.Contains(t, result.Content, "Status")
	assert.Contains(t, result.Content, "Description")
	assert.Contains(t, result.Content, "CreatedAt")
}

func TestListTool_InvalidJSON(t *testing.T) {
	_, _, _, list, _, _, _ := setup(t)
	result, err := list.Execute(context.Background(), json.RawMessage(`"not-object"`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestListTool_ContextCancel(t *testing.T) {
	_, _, _, list, _, _, _ := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := list.Execute(ctx, json.RawMessage(`{}`))
	assert.Error(t, err)
}
