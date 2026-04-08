package cron

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

// setup returns a fresh Manager and all four tool instances backed by it.
func setup(t *testing.T) (*Manager, CreateTool, DeleteTool, ListTool, ToggleTool) {
	t.Helper()
	mgr := NewManager()
	return mgr,
		CreateTool{Mgr: mgr},
		DeleteTool{Mgr: mgr},
		ListTool{Mgr: mgr},
		ToggleTool{Mgr: mgr}
}

// ---- helpers ----

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

// ---- CreateTool ----

func TestCreateTool_ValidJob(t *testing.T) {
	_, create, _, _, _ := setup(t)
	args := mustJSON(t, map[string]any{
		"name":     "backup",
		"schedule": "0 2 * * *",
		"command":  "pg_dump mydb",
	})
	result, err := create.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "backup")
}

func TestCreateTool_InvalidSchedule(t *testing.T) {
	_, create, _, _, _ := setup(t)
	args := mustJSON(t, map[string]any{
		"name":     "bad",
		"schedule": "not-cron",
		"command":  "echo",
	})
	result, err := create.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid schedule")
}

func TestCreateTool_DefaultEnabled(t *testing.T) {
	mgr, create, _, _, _ := setup(t)
	args := mustJSON(t, map[string]any{
		"name":     "default-en",
		"schedule": "0 * * * *",
		"command":  "echo",
	})
	result, err := create.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	job, ok := mgr.Get("default-en")
	require.True(t, ok)
	assert.True(t, job.Enabled)
}

func TestCreateTool_ExplicitDisabled(t *testing.T) {
	mgr, create, _, _, _ := setup(t)
	args := mustJSON(t, map[string]any{
		"name":     "disabled-job",
		"schedule": "0 * * * *",
		"command":  "echo",
		"enabled":  false,
	})
	result, err := create.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	job, ok := mgr.Get("disabled-job")
	require.True(t, ok)
	assert.False(t, job.Enabled)
}

func TestCreateTool_WithCwd(t *testing.T) {
	mgr, create, _, _, _ := setup(t)
	args := mustJSON(t, map[string]any{
		"name":     "with-cwd",
		"schedule": "0 * * * *",
		"command":  "ls",
		"cwd":      "/tmp",
	})
	result, err := create.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	job, ok := mgr.Get("with-cwd")
	require.True(t, ok)
	assert.Equal(t, "/tmp", job.Cwd)
}

// ---- DeleteTool ----

func TestDeleteTool_Existing(t *testing.T) {
	_, create, deleteTool, _, _ := setup(t)

	// Create first
	createArgs := mustJSON(t, map[string]any{
		"name": "to-delete", "schedule": "0 * * * *", "command": "echo",
	})
	_, err := create.Execute(context.Background(), createArgs)
	require.NoError(t, err)

	// Delete
	delArgs := mustJSON(t, map[string]any{"name": "to-delete"})
	result, err := deleteTool.Execute(context.Background(), delArgs)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "to-delete")
}

func TestDeleteTool_NotFound(t *testing.T) {
	_, _, deleteTool, _, _ := setup(t)
	args := mustJSON(t, map[string]any{"name": "ghost"})
	result, err := deleteTool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

// ---- ListTool ----

func TestListTool_Empty(t *testing.T) {
	_, _, _, list, _ := setup(t)
	result, err := list.Execute(context.Background(), json.RawMessage(`{}`))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "No cron jobs")
}

func TestListTool_ShowsJobs(t *testing.T) {
	_, create, _, list, _ := setup(t)

	createArgs := mustJSON(t, map[string]any{
		"name": "list-me", "schedule": "*/5 * * * *", "command": "date",
	})
	_, err := create.Execute(context.Background(), createArgs)
	require.NoError(t, err)

	result, err := list.Execute(context.Background(), json.RawMessage(`{}`))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "list-me")
	assert.Contains(t, result.Content, "*/5 * * * *")
	assert.Contains(t, result.Content, "date")
}

// ---- ToggleTool ----

func TestToggleTool_Enable(t *testing.T) {
	mgr, create, _, _, toggle := setup(t)

	createArgs := mustJSON(t, map[string]any{
		"name": "tog", "schedule": "0 * * * *", "command": "echo", "enabled": false,
	})
	_, err := create.Execute(context.Background(), createArgs)
	require.NoError(t, err)

	toggleArgs := mustJSON(t, map[string]any{"name": "tog", "enabled": true})
	result, err := toggle.Execute(context.Background(), toggleArgs)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "enabled")

	job, _ := mgr.Get("tog")
	assert.True(t, job.Enabled)
}

func TestToggleTool_Disable(t *testing.T) {
	mgr, create, _, _, toggle := setup(t)

	createArgs := mustJSON(t, map[string]any{
		"name": "tog2", "schedule": "0 * * * *", "command": "echo", "enabled": true,
	})
	_, err := create.Execute(context.Background(), createArgs)
	require.NoError(t, err)

	toggleArgs := mustJSON(t, map[string]any{"name": "tog2", "enabled": false})
	result, err := toggle.Execute(context.Background(), toggleArgs)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "disabled")

	job, _ := mgr.Get("tog2")
	assert.False(t, job.Enabled)
}

func TestToggleTool_NotFound(t *testing.T) {
	_, _, _, _, toggle := setup(t)
	args := mustJSON(t, map[string]any{"name": "missing", "enabled": true})
	result, err := toggle.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

// ---- Invalid JSON for each tool ----

func TestCreateTool_InvalidJSON(t *testing.T) {
	_, create, _, _, _ := setup(t)
	result, err := create.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestDeleteTool_InvalidJSON(t *testing.T) {
	_, _, deleteTool, _, _ := setup(t)
	result, err := deleteTool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestListTool_InvalidJSON(t *testing.T) {
	_, _, _, list, _ := setup(t)
	result, err := list.Execute(context.Background(), json.RawMessage(`"not-object"`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestToggleTool_InvalidJSON(t *testing.T) {
	_, _, _, _, toggle := setup(t)
	result, err := toggle.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

// ---- Context cancel for each tool ----

func TestCreateTool_ContextCancel(t *testing.T) {
	_, create, _, _, _ := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := mustJSON(t, map[string]any{
		"name": "ctx-test", "schedule": "0 * * * *", "command": "echo",
	})
	_, err := create.Execute(ctx, args)
	assert.Error(t, err)
}

func TestDeleteTool_ContextCancel(t *testing.T) {
	_, _, deleteTool, _, _ := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := mustJSON(t, map[string]any{"name": "x"})
	_, err := deleteTool.Execute(ctx, args)
	assert.Error(t, err)
}

func TestListTool_ContextCancel(t *testing.T) {
	_, _, _, list, _ := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := list.Execute(ctx, json.RawMessage(`{}`))
	assert.Error(t, err)
}

func TestToggleTool_ContextCancel(t *testing.T) {
	_, _, _, _, toggle := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := mustJSON(t, map[string]any{"name": "x", "enabled": true})
	_, err := toggle.Execute(ctx, args)
	assert.Error(t, err)
}

// ---- Create then List then Delete integration ----

func TestIntegration_CreateListDelete(t *testing.T) {
	mgr, create, deleteTool, list, _ := setup(t)

	// Create two jobs
	for _, j := range []map[string]any{
		{"name": "job-a", "schedule": "0 * * * *", "command": "echo a"},
		{"name": "job-b", "schedule": "0 0 * * *", "command": "echo b"},
	} {
		result, err := create.Execute(context.Background(), mustJSON(t, j))
		require.NoError(t, err)
		assert.False(t, result.IsError)
	}

	// List shows both
	listResult, err := list.Execute(context.Background(), json.RawMessage(`{}`))
	require.NoError(t, err)
	assert.False(t, listResult.IsError)
	assert.Contains(t, listResult.Content, "job-a")
	assert.Contains(t, listResult.Content, "job-b")

	// Delete one
	delResult, err := deleteTool.Execute(context.Background(), mustJSON(t, map[string]any{"name": "job-a"}))
	require.NoError(t, err)
	assert.False(t, delResult.IsError)

	// List now shows only the other
	listResult2, err := list.Execute(context.Background(), json.RawMessage(`{}`))
	require.NoError(t, err)
	assert.Contains(t, listResult2.Content, "job-b")
	assert.NotContains(t, listResult2.Content, "job-a")

	// Manager confirms
	assert.Len(t, mgr.List(), 1)
}

// Verify all four tools satisfy the interface.
var (
	_ tools.Tool = CreateTool{}
	_ tools.Tool = DeleteTool{}
	_ tools.Tool = ListTool{}
	_ tools.Tool = ToggleTool{}
)
