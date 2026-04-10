package memory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStore(t *testing.T) *Store {
	t.Helper()
	// Create a temp dir to use as cwd, but manually set the store dir.
	tmpDir := t.TempDir()
	memDir := filepath.Join(tmpDir, "memory-test")
	require.NoError(t, os.MkdirAll(memDir, 0o755))
	return &Store{dir: memDir}
}

func TestStore_Add(t *testing.T) {
	s := setupStore(t)

	path, err := s.Add("My Test Memory", "This is the content")
	require.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, filepath.Base(path), "my_test_memory")

	// File should exist.
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "This is the content\n", string(data))

	// MEMORY.md should reference it.
	ep := filepath.Join(s.dir, "MEMORY.md")
	indexData, err := os.ReadFile(ep)
	require.NoError(t, err)
	assert.Contains(t, string(indexData), "my_test_memory.md")
}

func TestStore_Remove(t *testing.T) {
	s := setupStore(t)

	path, err := s.Add("Delete Me", "content to remove")
	require.NoError(t, err)

	removed, err := s.Remove("delete_me")
	require.NoError(t, err)
	assert.True(t, removed)

	// File should be gone.
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestStore_Remove_Missing(t *testing.T) {
	s := setupStore(t)

	removed, err := s.Remove("nonexistent")
	require.NoError(t, err)
	assert.False(t, removed)
}

func TestStore_List(t *testing.T) {
	s := setupStore(t)

	_, err := s.Add("Alpha", "first")
	require.NoError(t, err)
	_, err = s.Add("Beta", "second")
	require.NoError(t, err)

	names, err := s.List()
	require.NoError(t, err)
	assert.Len(t, names, 2)
	assert.Equal(t, "alpha.md", names[0])
	assert.Equal(t, "beta.md", names[1])
}

func TestStore_LoadPrompt(t *testing.T) {
	s := setupStore(t)

	_, err := s.Add("Test", "memory content")
	require.NoError(t, err)

	prompt, err := s.LoadPrompt(0)
	require.NoError(t, err)
	assert.NotEmpty(t, prompt)
	assert.Contains(t, prompt, "Memory Index")
	assert.Contains(t, prompt, "test.md")
}

func TestStore_LoadPrompt_NotExist(t *testing.T) {
	s := setupStore(t)

	prompt, err := s.LoadPrompt(0)
	require.NoError(t, err)
	assert.Empty(t, prompt)
}
