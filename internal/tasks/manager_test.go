package tasks

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// waitForStatus polls until the task reaches the expected status or times out.
func waitForStatus(m *Manager, id string, status Status, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		rec, ok := m.Get(id)
		if ok && rec.Status == status {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func TestCreateShell_Echo(t *testing.T) {
	m := NewManager()
	rec, err := m.CreateShell(context.Background(), "echo hello", "echo test", "")
	require.NoError(t, err)
	require.NotNil(t, rec)

	assert.Equal(t, TypeLocalBash, rec.Type)
	assert.Equal(t, StatusRunning, rec.Status)
	assert.Equal(t, "echo test", rec.Description)
	assert.NotEmpty(t, rec.ID)
	assert.True(t, strings.HasPrefix(rec.ID, "b"), "task ID should start with 'b'")

	ok := waitForStatus(m, rec.ID, StatusCompleted, 5*time.Second)
	assert.True(t, ok, "task should complete")

	updated, found := m.Get(rec.ID)
	require.True(t, found)
	assert.Equal(t, StatusCompleted, updated.Status)

	output, err := m.ReadOutput(rec.ID, 0)
	require.NoError(t, err)
	assert.Contains(t, output, "hello")
}

func TestCreateShell_FailCommand(t *testing.T) {
	m := NewManager()
	rec, err := m.CreateShell(context.Background(), "exit 1", "failing command", "")
	require.NoError(t, err)
	require.NotNil(t, rec)

	ok := waitForStatus(m, rec.ID, StatusFailed, 5*time.Second)
	assert.True(t, ok, "task should fail")

	updated, found := m.Get(rec.ID)
	require.True(t, found)
	assert.Equal(t, StatusFailed, updated.Status)
	require.NotNil(t, updated.ReturnCode)
	assert.Equal(t, 1, *updated.ReturnCode)
}

func TestGet_Found(t *testing.T) {
	m := NewManager()
	rec, err := m.CreateShell(context.Background(), "echo test", "", "")
	require.NoError(t, err)

	found, ok := m.Get(rec.ID)
	assert.True(t, ok)
	assert.Equal(t, rec.ID, found.ID)
}

func TestGet_Missing(t *testing.T) {
	m := NewManager()
	rec, ok := m.Get("nonexistent")
	assert.Nil(t, rec)
	assert.False(t, ok)
}

func TestList_All(t *testing.T) {
	m := NewManager()
	rec1, err := m.CreateShell(context.Background(), "echo first", "task 1", "")
	require.NoError(t, err)

	// Small delay to ensure different CreatedAt timestamps.
	time.Sleep(50 * time.Millisecond)

	rec2, err := m.CreateShell(context.Background(), "echo second", "task 2", "")
	require.NoError(t, err)

	// Wait for both to complete.
	waitForStatus(m, rec1.ID, StatusCompleted, 5*time.Second)
	waitForStatus(m, rec2.ID, StatusCompleted, 5*time.Second)

	all := m.List("")
	assert.Len(t, all, 2)
	// Sorted by CreatedAt descending (newest first).
	assert.Equal(t, rec2.ID, all[0].ID)
	assert.Equal(t, rec1.ID, all[1].ID)
}

func TestList_FilterStatus(t *testing.T) {
	m := NewManager()

	rec1, err := m.CreateShell(context.Background(), "echo ok", "", "")
	require.NoError(t, err)
	rec2, err := m.CreateShell(context.Background(), "exit 1", "", "")
	require.NoError(t, err)

	waitForStatus(m, rec1.ID, StatusCompleted, 5*time.Second)
	waitForStatus(m, rec2.ID, StatusFailed, 5*time.Second)

	completed := m.List(StatusCompleted)
	for _, r := range completed {
		assert.Equal(t, StatusCompleted, r.Status)
	}

	failed := m.List(StatusFailed)
	for _, r := range failed {
		assert.Equal(t, StatusFailed, r.Status)
	}
}

func TestStop_KillsProcess(t *testing.T) {
	m := NewManager()
	rec, err := m.CreateShell(context.Background(), "sleep 30", "long sleep", "")
	require.NoError(t, err)

	assert.Equal(t, StatusRunning, rec.Status)

	err = m.Stop(context.Background(), rec.ID)
	require.NoError(t, err)

	updated, found := m.Get(rec.ID)
	require.True(t, found)
	assert.Equal(t, StatusKilled, updated.Status)
	assert.NotNil(t, updated.EndedAt)
}

func TestStop_AlreadyCompleted(t *testing.T) {
	m := NewManager()
	rec, err := m.CreateShell(context.Background(), "echo done", "", "")
	require.NoError(t, err)

	waitForStatus(m, rec.ID, StatusCompleted, 5*time.Second)

	// Stopping an already completed task should not error.
	err = m.Stop(context.Background(), rec.ID)
	assert.NoError(t, err)
}

func TestStop_NonexistentTask(t *testing.T) {
	m := NewManager()
	err := m.Stop(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestReadOutput(t *testing.T) {
	m := NewManager()
	rec, err := m.CreateShell(context.Background(), "echo test_output_content", "", "")
	require.NoError(t, err)

	waitForStatus(m, rec.ID, StatusCompleted, 5*time.Second)

	output, err := m.ReadOutput(rec.ID, 0)
	require.NoError(t, err)
	assert.Contains(t, output, "test_output_content")
}

func TestReadOutput_Truncate(t *testing.T) {
	m := NewManager()
	// Generate more than 100 bytes of output.
	rec, err := m.CreateShell(context.Background(), "seq 1 50", "", "")
	require.NoError(t, err)

	waitForStatus(m, rec.ID, StatusCompleted, 5*time.Second)

	output, err := m.ReadOutput(rec.ID, 50)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(output), 50)
}

func TestReadOutput_NonexistentTask(t *testing.T) {
	m := NewManager()
	_, err := m.ReadOutput("nonexistent", 0)
	assert.Error(t, err)
}

func TestUpdate(t *testing.T) {
	m := NewManager()
	rec, err := m.CreateShell(context.Background(), "echo test", "original", "")
	require.NoError(t, err)

	updated, err := m.Update(rec.ID, "updated description", 50, "halfway done")
	require.NoError(t, err)
	assert.Equal(t, "updated description", updated.Description)
	assert.Equal(t, "50", updated.Metadata["progress"])
	assert.Equal(t, "halfway done", updated.Metadata["statusNote"])
}

func TestUpdate_NonexistentTask(t *testing.T) {
	m := NewManager()
	_, err := m.Update("nonexistent", "desc", 0, "")
	assert.Error(t, err)
}

func TestTaskID_Format(t *testing.T) {
	id := taskID(TypeLocalBash)
	assert.True(t, strings.HasPrefix(id, "b"), "bash task ID should start with 'b'")
	assert.Len(t, id, 9) // "b" + 8 hex chars

	id2 := taskID(TypeLocalAgent)
	assert.True(t, strings.HasPrefix(id2, "a"), "agent task ID should start with 'a'")
	assert.Len(t, id2, 9)
}

func TestTaskID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := taskID(TypeLocalBash)
		assert.False(t, ids[id], "duplicate task ID generated: %s", id)
		ids[id] = true
	}
}
