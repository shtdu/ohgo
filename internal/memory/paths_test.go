package memory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectDir(t *testing.T) {
	cwd := t.TempDir()
	dir, err := ProjectDir(cwd)
	require.NoError(t, err)

	// Project memory lives inside CWD/.ohgo/data/memory/
	abs, _ := filepath.Abs(cwd)
	expected := filepath.Join(abs, ".ohgo", "data", "memory")
	assert.Equal(t, expected, dir)

	// Directory must exist
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestEntrypoint(t *testing.T) {
	cwd := t.TempDir()
	ep, err := Entrypoint(cwd)
	require.NoError(t, err)
	assert.Equal(t, "MEMORY.md", filepath.Base(ep))

	// Entrypoint must be inside CWD/.ohgo/data/memory/
	abs, _ := filepath.Abs(cwd)
	expected := filepath.Join(abs, ".ohgo", "data", "memory", "MEMORY.md")
	assert.Equal(t, expected, ep)
}

func TestProjectDir_ResolvesRelative(t *testing.T) {
	dir, err := ProjectDir(".")
	require.NoError(t, err)
	abs, _ := filepath.Abs(".")
	expected := filepath.Join(abs, ".ohgo", "data", "memory")
	assert.Equal(t, expected, dir)
}

func TestProjectDir_SameCWD_SamePath(t *testing.T) {
	cwd := t.TempDir()
	dir1, err := ProjectDir(cwd)
	require.NoError(t, err)
	dir2, err := ProjectDir(cwd)
	require.NoError(t, err)
	assert.Equal(t, dir1, dir2)
}

func TestProjectDir_DifferentCWD_DifferentPath(t *testing.T) {
	cwd1 := t.TempDir()
	cwd2 := t.TempDir()
	dir1, err := ProjectDir(cwd1)
	require.NoError(t, err)
	dir2, err := ProjectDir(cwd2)
	require.NoError(t, err)
	assert.NotEqual(t, dir1, dir2)
}

func TestPersonalDir(t *testing.T) {
	dir, err := PersonalDir()
	require.NoError(t, err)
	assert.NotEmpty(t, dir)
	assert.Contains(t, dir, "memory")

	// Must be under the global data dir (not inside any CWD)
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
