//go:build integration

package memory_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/memory"
)

// EARS: REQ-MC-001
func TestIntegration_Memory_PersistenceAcrossInstances(t *testing.T) {
	dir := t.TempDir()

	// Create first store and add memory
	store1, err := memory.NewStore(dir)
	require.NoError(t, err)

	path, err := store1.Add("Test Memory", "This is test content")
	require.NoError(t, err)
	assert.NotEmpty(t, path)

	// Create second store for the same project — should see the memory
	store2, err := memory.NewStore(dir)
	require.NoError(t, err)

	files, err := store2.List()
	require.NoError(t, err)
	assert.Contains(t, files, "test_memory.md", "memory should persist across store instances")
}

// EARS: REQ-MC-005
func TestIntegration_Memory_AddRemoveUpdatesIndex(t *testing.T) {
	dir := t.TempDir()
	store, err := memory.NewStore(dir)
	require.NoError(t, err)

	// Add entry
	_, err = store.Add("My Entry", "Some content here")
	require.NoError(t, err)

	// Verify MEMORY.md index contains the entry
	indexData, err := os.ReadFile(filepath.Join(store.ProjectDir(), "MEMORY.md"))
	require.NoError(t, err)
	assert.Contains(t, string(indexData), "my_entry.md")
	assert.Contains(t, string(indexData), "My Entry")

	// Remove entry
	removed, err := store.Remove("my_entry")
	require.NoError(t, err)
	assert.True(t, removed)

	// Verify removed from list
	files, err := store.List()
	require.NoError(t, err)
	assert.NotContains(t, files, "my_entry.md")

	// Verify MEMORY.md index updated
	indexData, err = os.ReadFile(filepath.Join(store.ProjectDir(), "MEMORY.md"))
	require.NoError(t, err)
	assert.NotContains(t, string(indexData), "my_entry.md")
}

// EARS: REQ-MC-002
func TestIntegration_Memory_LoadPrompt(t *testing.T) {
	dir := t.TempDir()
	store, err := memory.NewStore(dir)
	require.NoError(t, err)

	// Add entries
	_, err = store.Add("First Note", "First content")
	require.NoError(t, err)
	_, err = store.Add("Second Note", "Second content")
	require.NoError(t, err)

	// Load prompt should return content from MEMORY.md
	prompt, err := store.LoadPrompt(100)
	require.NoError(t, err)
	assert.Contains(t, prompt, "first_note")
	assert.Contains(t, prompt, "second_note")
}

// EARS: REQ-MC-002
func TestIntegration_Memory_LoadPromptEmpty(t *testing.T) {
	dir := t.TempDir()
	store, err := memory.NewStore(dir)
	require.NoError(t, err)

	// No memories added — prompt should have no memory content
	prompt, err := store.LoadPrompt(100)
	require.NoError(t, err)
	// Even with no memories, LoadPrompt may produce section headers
	assert.True(t, len(prompt) < 200, "empty store should produce minimal prompt, got: %q", prompt)
}

// EARS: REQ-MC-005
func TestIntegration_Memory_RemoveNonExistent(t *testing.T) {
	dir := t.TempDir()
	store, err := memory.NewStore(dir)
	require.NoError(t, err)

	removed, err := store.Remove("nonexistent")
	require.NoError(t, err)
	assert.False(t, removed, "removing nonexistent entry should return false")
}

// EARS: REQ-MC-001
func TestIntegration_Memory_PersonalScope(t *testing.T) {
	dir := t.TempDir()
	store, err := memory.NewStore(dir)
	require.NoError(t, err)

	// Add personal memory
	path, err := store.AddPersonal("Personal Note", "personal content")
	require.NoError(t, err)
	assert.NotEmpty(t, path)

	// List personal should contain it
	files, err := store.ListPersonal()
	require.NoError(t, err)
	assert.Contains(t, files, "personal_note.md")

	// Remove personal
	removed, err := store.RemovePersonal("personal_note")
	require.NoError(t, err)
	assert.True(t, removed)
}
