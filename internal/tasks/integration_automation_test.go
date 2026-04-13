//go:build integration

package tasks_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/tasks"
)

// EARS: REQ-AT-001
func TestIntegration_Task_ShellExecution(t *testing.T) {
	mgr := tasks.NewManager()

	rec, err := mgr.CreateShell(context.Background(), "echo hello task", "test echo", t.TempDir())
	require.NoError(t, err)
	assert.NotEmpty(t, rec.ID)
	assert.Equal(t, tasks.StatusRunning, rec.Status)

	// Wait for completion
	time.Sleep(500 * time.Millisecond)

	updated, found := mgr.Get(rec.ID)
	require.True(t, found)
	assert.Equal(t, tasks.StatusCompleted, updated.Status)
}

// EARS: REQ-AT-002
func TestIntegration_Task_LifecycleStates(t *testing.T) {
	mgr := tasks.NewManager()

	// Create: should be running
	rec, err := mgr.CreateShell(context.Background(), "sleep 1", "test lifecycle", t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, tasks.StatusRunning, rec.Status)
	assert.Nil(t, rec.EndedAt)

	// Wait for completion
	time.Sleep(1500 * time.Millisecond)

	// Get: should be completed
	final, found := mgr.Get(rec.ID)
	require.True(t, found)
	assert.Equal(t, tasks.StatusCompleted, final.Status)
	assert.NotNil(t, final.EndedAt)
}

// EARS: REQ-AT-002
func TestIntegration_Task_List(t *testing.T) {
	mgr := tasks.NewManager()

	_, err := mgr.CreateShell(context.Background(), "echo a", "task a", t.TempDir())
	require.NoError(t, err)
	_, err = mgr.CreateShell(context.Background(), "echo b", "task b", t.TempDir())
	require.NoError(t, err)

	// List all
	all := mgr.List("")
	assert.GreaterOrEqual(t, len(all), 2)

	// List running only
	time.Sleep(500 * time.Millisecond)
	running := mgr.List(tasks.StatusRunning)
	// After 500ms, echo commands likely completed
	assert.LessOrEqual(t, len(running), 2)
}

// EARS: REQ-AT-004
func TestIntegration_Task_OutputRetrieval(t *testing.T) {
	mgr := tasks.NewManager()

	rec, err := mgr.CreateShell(context.Background(), "echo output_test_123", "test output", t.TempDir())
	require.NoError(t, err)

	// Wait for completion
	time.Sleep(500 * time.Millisecond)

	output, err := mgr.ReadOutput(rec.ID, 0)
	require.NoError(t, err)
	assert.Contains(t, output, "output_test_123")
}

// EARS: REQ-AT-005
func TestIntegration_Task_ProgressUpdate(t *testing.T) {
	mgr := tasks.NewManager()

	rec, err := mgr.CreateShell(context.Background(), "echo progress", "test progress", t.TempDir())
	require.NoError(t, err)

	updated, err := mgr.Update(rec.ID, "updated description", 50, "halfway")
	require.NoError(t, err)
	assert.Equal(t, "updated description", updated.Description)

	// Verify persisted
	got, found := mgr.Get(rec.ID)
	require.True(t, found)
	assert.Equal(t, "updated description", got.Description)
}

// EARS: REQ-AT-001
func TestIntegration_Task_FailedCommand(t *testing.T) {
	mgr := tasks.NewManager()

	rec, err := mgr.CreateShell(context.Background(), "exit 1", "failing task", t.TempDir())
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	final, found := mgr.Get(rec.ID)
	require.True(t, found)
	assert.Equal(t, tasks.StatusFailed, final.Status)
}

// EARS: REQ-AT-004
func TestIntegration_Task_LargeOutput(t *testing.T) {
	mgr := tasks.NewManager()

	// Generate large output
	rec, err := mgr.CreateShell(context.Background(),
		"for i in $(seq 1 100); do echo \"line $i - some content here\"; done",
		"large output", t.TempDir(),
	)
	require.NoError(t, err)

	time.Sleep(1000 * time.Millisecond)

	output, err := mgr.ReadOutput(rec.ID, 0)
	require.NoError(t, err)
	assert.True(t, strings.Count(output, "line") >= 50, "should have many lines")
}
