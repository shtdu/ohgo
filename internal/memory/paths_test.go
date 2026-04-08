package memory

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectDir(t *testing.T) {
	dir, err := ProjectDir(t.TempDir())
	require.NoError(t, err)
	assert.NotEmpty(t, dir)
	assert.Contains(t, dir, "memory")
}

func TestEntrypoint(t *testing.T) {
	ep, err := Entrypoint(t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, "MEMORY.md", filepath.Base(ep))
}
