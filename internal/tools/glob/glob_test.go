package glob

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// Create test file structure:
	//   a.go, b.txt, sub/c.go, sub/d.txt, sub/deep/e.go, .git/HEAD
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte("go"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.txt"), []byte("txt"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "c.go"), []byte("go"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "d.txt"), []byte("txt"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub", "deep"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "deep", "e.go"), []byte("go"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref"), 0644))
	return dir
}

func TestGlobTool_Name(t *testing.T) {
	assert.Equal(t, "glob", GlobTool{}.Name())
}

func TestGlobTool_RecursiveGo(t *testing.T) {
	dir := setupTestDir(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	tool := GlobTool{}
	args, _ := json.Marshal(map[string]string{"pattern": "**/*.go"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	matches := strings.Split(strings.TrimSpace(result.Content), "\n")
	sort.Strings(matches)
	assert.Contains(t, matches, "a.go")
	assert.Contains(t, matches, filepath.Join("sub", "c.go"))
	assert.Contains(t, matches, filepath.Join("sub", "deep", "e.go"))
	assert.NotContains(t, result.Content, ".git")
}

func TestGlobTool_SimpleStar(t *testing.T) {
	dir := setupTestDir(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	tool := GlobTool{}
	args, _ := json.Marshal(map[string]string{"pattern": "*.go"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "a.go")
	assert.NotContains(t, result.Content, "b.txt")
}

func TestGlobTool_NoMatches(t *testing.T) {
	dir := setupTestDir(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	tool := GlobTool{}
	args, _ := json.Marshal(map[string]string{"pattern": "*.xyz"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "(no matches)", result.Content)
}

func TestGlobTool_LimitExceeded(t *testing.T) {
	dir := setupTestDir(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	tool := GlobTool{}
	args, _ := json.Marshal(map[string]any{"pattern": "**/*", "limit": 2})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	matches := strings.Split(strings.TrimSpace(result.Content), "\n")
	assert.LessOrEqual(t, len(matches), 2)
}

func TestGlobTool_InvalidJSON(t *testing.T) {
	tool := GlobTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestGlobTool_EmptyPattern(t *testing.T) {
	tool := GlobTool{}
	args, _ := json.Marshal(map[string]string{"pattern": ""})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

var _ tools.Tool = GlobTool{}
