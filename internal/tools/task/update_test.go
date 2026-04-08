package task

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateTool_Description(t *testing.T) {
	mgr, _, _, _, _, _, update := setup(t)
	id := mustCreateTask(t, mgr, "echo hello", "original description")

	args := mustJSON(t, map[string]any{
		"task_id":     id,
		"description": "updated description",
	})
	result, err := update.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "updated description")

	// Verify via manager directly.
	rec, found := mgr.Get(id)
	require.True(t, found)
	assert.Equal(t, "updated description", rec.Description)
}

func TestUpdateTool_Progress(t *testing.T) {
	mgr, _, _, _, _, _, update := setup(t)
	id := mustCreateTask(t, mgr, "echo hello", "progress test")

	progress := 75
	args := mustJSON(t, map[string]any{
		"task_id":  id,
		"progress": progress,
	})
	result, err := update.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	// Verify via manager directly.
	rec, found := mgr.Get(id)
	require.True(t, found)
	assert.Equal(t, "75", rec.Metadata["progress"])
}

func TestUpdateTool_StatusNote(t *testing.T) {
	mgr, _, _, _, _, _, update := setup(t)
	id := mustCreateTask(t, mgr, "echo hello", "note test")

	args := mustJSON(t, map[string]any{
		"task_id":     id,
		"status_note": "deploying to staging",
	})
	result, err := update.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "deploying to staging")

	// Verify via manager directly.
	rec, found := mgr.Get(id)
	require.True(t, found)
	assert.Equal(t, "deploying to staging", rec.Metadata["statusNote"])
}

func TestUpdateTool_MultipleFields(t *testing.T) {
	mgr, _, _, _, _, _, update := setup(t)
	id := mustCreateTask(t, mgr, "echo hello", "multi update")

	args := mustJSON(t, map[string]any{
		"task_id":     id,
		"description": "new desc",
		"progress":    50,
		"status_note": "halfway there",
	})
	result, err := update.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	rec, found := mgr.Get(id)
	require.True(t, found)
	assert.Equal(t, "new desc", rec.Description)
	assert.Equal(t, "50", rec.Metadata["progress"])
	assert.Equal(t, "halfway there", rec.Metadata["statusNote"])
}

func TestUpdateTool_NotFound(t *testing.T) {
	_, _, _, _, _, _, update := setup(t)
	args := mustJSON(t, map[string]any{
		"task_id":     "nonexistent",
		"description": "ghost update",
	})
	result, err := update.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

func TestUpdateTool_EmptyTaskID(t *testing.T) {
	_, _, _, _, _, _, update := setup(t)
	args := mustJSON(t, map[string]any{
		"task_id":     "",
		"description": "no id",
	})
	result, err := update.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "task_id is required")
}

func TestUpdateTool_InvalidJSON(t *testing.T) {
	_, _, _, _, _, _, update := setup(t)
	result, err := update.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestUpdateTool_ContextCancel(t *testing.T) {
	_, _, _, _, _, _, update := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args := mustJSON(t, map[string]any{"task_id": "x"})
	_, err := update.Execute(ctx, args)
	assert.Error(t, err)
}
