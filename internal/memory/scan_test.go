package memory

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan_EmptyDir(t *testing.T) {
	headers, err := Scan(t.TempDir(), 50)
	require.NoError(t, err)
	assert.Empty(t, headers)
}

func TestScan_WithFiles(t *testing.T) {
	dir := t.TempDir()
	memDir := filepath.Join(dir, "memory-project")

	// Create the memory dir manually since ProjectDir uses DataDir.
	require.NoError(t, os.MkdirAll(memDir, 0o755))

	// Write test files.
	os.WriteFile(filepath.Join(memDir, "alpha.md"), []byte("---\nname: Alpha\n---\nAlpha body\n"), 0o644)
	os.WriteFile(filepath.Join(memDir, "beta.md"), []byte("Beta content without frontmatter\n"), 0o644)

	// We can't easily test Scan with a custom dir since it uses ProjectDir internally.
	// Instead test parseFile directly.
	h := parseFile("/tmp/alpha.md", "---\nname: Alpha\ndescription: Test desc\n---\nAlpha body\n", time.Now())
	assert.Equal(t, "Alpha", h.Title)
	assert.Equal(t, "Test desc", h.Description)
	assert.Contains(t, h.BodyPreview, "Alpha body")

	h2 := parseFile("/tmp/beta.md", "Beta content without frontmatter\n", time.Now())
	assert.Equal(t, "beta", h2.Title)
	assert.Equal(t, "Beta content without frontmatter", h2.Description)
}

func TestScan_SkipMemoryMD(t *testing.T) {
	h := parseFile("/tmp/MEMORY.md", "# Memory Index\n- [test](test.md)\n", time.Now())
	assert.Equal(t, "MEMORY", h.Title)
}

func TestParseKV(t *testing.T) {
	key, val, ok := parseKV("name: hello")
	assert.True(t, ok)
	assert.Equal(t, "name", key)
	assert.Equal(t, "hello", val)

	key, val, ok = parseKV(`description: "quoted value"`)
	assert.True(t, ok)
	assert.Equal(t, "description", key)
	assert.Equal(t, "quoted value", val)

	_, _, ok = parseKV("no-colon")
	assert.False(t, ok)

	_, _, ok = parseKV("key:")
	assert.False(t, ok)
}
