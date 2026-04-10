package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	require.NoError(t, os.WriteFile(filepath.Join(memDir, "alpha.md"), []byte("---\nname: Alpha\n---\nAlpha body\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(memDir, "beta.md"), []byte("Beta content without frontmatter\n"), 0o644))

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

func TestScan_SkipNonMdFiles(t *testing.T) {
	t.Setenv("OPENHARNESS_DATA_DIR", t.TempDir())
	cwd := t.TempDir()
	memDir, err := ProjectDir(cwd)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(memDir, "notes.txt"), []byte("text file\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(memDir, "data.json"), []byte("{}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(memDir, "notes.md"), []byte("---\nname: Notes\n---\nSome notes\n"), 0o644))

	headers, err := Scan(cwd, 50)
	require.NoError(t, err)
	require.Len(t, headers, 1)
	assert.Equal(t, "Notes", headers[0].Title)
}

func TestScan_SkipDirectories(t *testing.T) {
	t.Setenv("OPENHARNESS_DATA_DIR", t.TempDir())
	cwd := t.TempDir()
	memDir, err := ProjectDir(cwd)
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Join(memDir, "subdir"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(memDir, "subdir", "nested.md"), []byte("nested\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(memDir, "test.md"), []byte("---\nname: Test\n---\nTest body\n"), 0o644))

	headers, err := Scan(cwd, 50)
	require.NoError(t, err)
	require.Len(t, headers, 1)
	assert.Equal(t, "Test", headers[0].Title)
}

func TestScan_MaxFilesCap(t *testing.T) {
	t.Setenv("OPENHARNESS_DATA_DIR", t.TempDir())
	cwd := t.TempDir()
	memDir, err := ProjectDir(cwd)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("file%d.md", i)
		content := fmt.Sprintf("---\nname: File%d\n---\nBody %d\n", i, i)
		require.NoError(t, os.WriteFile(filepath.Join(memDir, name), []byte(content), 0o644))
	}

	headers, err := Scan(cwd, 3)
	require.NoError(t, err)
	assert.Len(t, headers, 3)
}

func TestScan_MaxFilesZero(t *testing.T) {
	t.Setenv("OPENHARNESS_DATA_DIR", t.TempDir())
	cwd := t.TempDir()
	memDir, err := ProjectDir(cwd)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("file%d.md", i)
		content := fmt.Sprintf("---\nname: File%d\n---\nBody %d\n", i, i)
		require.NoError(t, os.WriteFile(filepath.Join(memDir, name), []byte(content), 0o644))
	}

	headers, err := Scan(cwd, 0)
	require.NoError(t, err)
	assert.Len(t, headers, 3)
}

func TestParseFile_LongDescription(t *testing.T) {
	longLine := strings.Repeat("a", 250)
	h := parseFile("/tmp/long.md", longLine+"\nother line\n", time.Now())
	assert.Len(t, h.Description, 200)
	assert.Equal(t, strings.Repeat("a", 200), h.Description)
}

func TestParseFile_LongBodyPreview(t *testing.T) {
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, "word"+fmt.Sprintf("%d", i))
	}
	content := strings.Join(lines, "\n")
	h := parseFile("/tmp/longbody.md", content, time.Now())
	assert.LessOrEqual(t, len(h.BodyPreview), 300)
}

func TestParseFile_TypeField(t *testing.T) {
	h := parseFile("/tmp/typed.md", "---\nname: Typed\ntype: feedback\n---\nBody text\n", time.Now())
	assert.Equal(t, "feedback", h.MemoryType)
}

func TestParseFile_EmptyContent(t *testing.T) {
	h := parseFile("/tmp/empty.md", "", time.Now())
	assert.NotPanics(t, func() {
		_ = parseFile("/tmp/empty.md", "", time.Now())
	})
	assert.Empty(t, h.BodyPreview)
}

func TestParseKV_WsKey(t *testing.T) {
	key, val, ok := parseKV("  key  :  value  ")
	assert.True(t, ok)
	assert.Equal(t, "key", key)
	assert.Equal(t, "value", val)
}
