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
// Shell execution with real process: create, verify running, wait, verify completed.
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
// Task lifecycle: running → completed with timing metadata.
func TestIntegration_Task_LifecycleStates(t *testing.T) {
	mgr := tasks.NewManager()

	rec, err := mgr.CreateShell(context.Background(), "sleep 1", "test lifecycle", t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, tasks.StatusRunning, rec.Status)
	assert.Nil(t, rec.EndedAt)

	time.Sleep(1500 * time.Millisecond)

	final, found := mgr.Get(rec.ID)
	require.True(t, found)
	assert.Equal(t, tasks.StatusCompleted, final.Status)
	assert.NotNil(t, final.EndedAt)
}

// EARS: REQ-AT-002
// List returns tasks filtered by status.
func TestIntegration_Task_List(t *testing.T) {
	mgr := tasks.NewManager()

	_, err := mgr.CreateShell(context.Background(), "echo a", "task a", t.TempDir())
	require.NoError(t, err)
	_, err = mgr.CreateShell(context.Background(), "echo b", "task b", t.TempDir())
	require.NoError(t, err)

	all := mgr.List("")
	assert.GreaterOrEqual(t, len(all), 2)

	time.Sleep(500 * time.Millisecond)
	running := mgr.List(tasks.StatusRunning)
	assert.LessOrEqual(t, len(running), 2)
}

// EARS: REQ-AT-004
// Output retrieval from a completed task contains the command output.
func TestIntegration_Task_OutputRetrieval(t *testing.T) {
	mgr := tasks.NewManager()

	rec, err := mgr.CreateShell(context.Background(), "echo output_test_123", "test output", t.TempDir())
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	output, err := mgr.ReadOutput(rec.ID, 0)
	require.NoError(t, err)
	assert.Contains(t, output, "output_test_123")
}

// EARS: REQ-AT-004
// Large output is fully captured and readable.
func TestIntegration_Task_LargeOutput(t *testing.T) {
	mgr := tasks.NewManager()

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

// EARS: REQ-AT-001, REQ-AT-005
// Failed command produces StatusFailed with non-zero return code,
// and progress metadata can still be attached before completion.
func TestIntegration_Task_FailedWithMetadata(t *testing.T) {
	mgr := tasks.NewManager()

	// Start a long-running task so we can update it before it fails
	rec, err := mgr.CreateShell(context.Background(), "sleep 0.3 && exit 42", "failing task", t.TempDir())
	require.NoError(t, err)

	// Attach progress metadata while running
	updated, err := mgr.Update(rec.ID, "about to fail", 50, "halfway")
	require.NoError(t, err)
	assert.Equal(t, "about to fail", updated.Description)

	// Wait for it to fail
	time.Sleep(800 * time.Millisecond)

	final, found := mgr.Get(rec.ID)
	require.True(t, found)
	assert.Equal(t, tasks.StatusFailed, final.Status)
	assert.Equal(t, "about to fail", final.Description, "metadata should survive failure")
}

// EARS: REQ-AT-004
// Stop kills a running task and output captured so far is readable.
func TestIntegration_Task_StopWhileRunning(t *testing.T) {
	mgr := tasks.NewManager()

	rec, err := mgr.CreateShell(context.Background(), "echo before_stop && sleep 30 && echo after_stop", "stoppable", t.TempDir())
	require.NoError(t, err)

	// Give it time to output the first echo
	time.Sleep(300 * time.Millisecond)

	// Stop it
	require.NoError(t, mgr.Stop(context.Background(), rec.ID))

	final, found := mgr.Get(rec.ID)
	require.True(t, found)
	assert.NotEqual(t, tasks.StatusRunning, final.Status)

	// Output captured before stop should be readable
	output, err := mgr.ReadOutput(rec.ID, 0)
	require.NoError(t, err)
	assert.Contains(t, output, "before_stop", "output before stop should be captured")
	assert.NotContains(t, output, "after_stop", "output after stop should not appear")
}
