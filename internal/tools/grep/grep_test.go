package grep

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

func setupGrepDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.go"), []byte("package main\n\nfunc main() {}\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.go"), []byte("package foo\n\nfunc helper() {}\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "c.txt"), []byte("Hello World\nhello world\n"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "d.go"), []byte("package sub\nfunc sub() {}\n"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".git"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".git", "config"), []byte("func gitignore\n"), 0644))
	return dir
}

func TestGrepTool_Name(t *testing.T) {
	assert.Equal(t, "grep", GrepTool{}.Name())
}

func TestGrepTool_PatternMatch(t *testing.T) {
	dir := setupGrepDir(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	tool := GrepTool{}
	args, _ := json.Marshal(map[string]string{"pattern": "func main"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "a.go:3:func main() {}")
	assert.NotContains(t, result.Content, ".git")
}

func TestGrepTool_CaseInsensitive(t *testing.T) {
	dir := setupGrepDir(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	tool := GrepTool{}
	args, _ := json.Marshal(map[string]any{"pattern": "hello", "case_sensitive": false})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	lines := strings.Split(result.Content, "\n")
	assert.GreaterOrEqual(t, len(lines), 2)
}

func TestGrepTool_FileFilter(t *testing.T) {
	dir := setupGrepDir(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	tool := GrepTool{}
	args, _ := json.Marshal(map[string]any{"pattern": "func", "glob": "*.go"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "a.go")
	assert.Contains(t, result.Content, "b.go")
	assert.NotContains(t, result.Content, "c.txt")
}

func TestGrepTool_NoMatches(t *testing.T) {
	dir := setupGrepDir(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	tool := GrepTool{}
	args, _ := json.Marshal(map[string]string{"pattern": "nonexistent_pattern_xyz"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "(no matches)", result.Content)
}

func TestGrepTool_SingleFile(t *testing.T) {
	dir := setupGrepDir(t)
	tool := GrepTool{}
	args, _ := json.Marshal(map[string]any{"pattern": "func", "path": filepath.Join(dir, "a.go")})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "func main")
}

func TestGrepTool_InvalidRegex(t *testing.T) {
	tool := GrepTool{}
	args, _ := json.Marshal(map[string]string{"pattern": "[invalid"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid regex")
}

func TestGrepTool_InvalidJSON(t *testing.T) {
	tool := GrepTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestGrepTool_RecursiveSearch(t *testing.T) {
	dir := setupGrepDir(t)
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	tool := GrepTool{}
	args, _ := json.Marshal(map[string]string{"pattern": "package"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "a.go")
	assert.Contains(t, result.Content, "b.go")
	assert.Contains(t, result.Content, filepath.Join("sub", "d.go"))
}

var _ tools.Tool = GrepTool{}
