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
	// Create a temp dir to use as both personal and project memory dirs.
	tmpDir := t.TempDir()
	projDir := filepath.Join(tmpDir, "project")
	persDir := filepath.Join(tmpDir, "personal")
	require.NoError(t, os.MkdirAll(projDir, 0o755))
	require.NoError(t, os.MkdirAll(persDir, 0o755))
	return &Store{projectDir: projDir, personalDir: persDir}
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
	ep := filepath.Join(s.projectDir, "MEMORY.md")
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

func TestStore_LoadPrompt_BothLayers(t *testing.T) {
	s := setupStore(t)

	_, err := s.AddPersonal("My Prefs", "I prefer table-driven tests")
	require.NoError(t, err)
	_, err = s.Add("Auth Rewrite", "Compliance-driven refactor")
	require.NoError(t, err)

	prompt, err := s.LoadPrompt(0)
	require.NoError(t, err)
	assert.Contains(t, prompt, "Personal Memory")
	assert.Contains(t, prompt, "my_prefs.md")
	assert.Contains(t, prompt, "Project Memory")
	assert.Contains(t, prompt, "auth_rewrite.md")
}

func TestStore_PersonalAddRemove(t *testing.T) {
	s := setupStore(t)

	path, err := s.AddPersonal("User Role", "I am a backend engineer")
	require.NoError(t, err)
	assert.Contains(t, filepath.Base(path), "user_role")

	names, err := s.ListPersonal()
	require.NoError(t, err)
	require.Len(t, names, 1)
	assert.Equal(t, "user_role.md", names[0])

	removed, err := s.RemovePersonal("user_role")
	require.NoError(t, err)
	assert.True(t, removed)

	names, err = s.ListPersonal()
	require.NoError(t, err)
	assert.Empty(t, names)
}

func TestStore_LoadPrompt_MaxLines(t *testing.T) {
	s := setupStore(t)

	_, err := s.Add("Test", "content")
	require.NoError(t, err)

	prompt, err := s.LoadPrompt(1)
	require.NoError(t, err)
	lines := 0
	for _, ch := range prompt {
		if ch == '\n' {
			lines++
		}
	}
	assert.LessOrEqual(t, lines, 1)
}
