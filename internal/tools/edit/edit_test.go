package edit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestEditTool_NameAndSchema(t *testing.T) {
	tool := EditTool{}
	assert.Equal(t, "edit_file", tool.Name())
	assert.Contains(t, tool.Description(), "string")

	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
	required := schema["required"].([]string)
	assert.Contains(t, required, "path")
	assert.Contains(t, required, "old_str")
	assert.Contains(t, required, "new_str")
}

func TestEditTool_ExactSingleMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	err := os.WriteFile(path, []byte("hello world\n"), 0644)
	require.NoError(t, err)

	tool := EditTool{}
	args, _ := json.Marshal(map[string]string{
		"path":    path,
		"old_str": "world",
		"new_str": "Go",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "Updated "+path, result.Content)

	// Verify file content
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "hello Go\n", string(data))
}

func TestEditTool_NoMatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	err := os.WriteFile(path, []byte("hello world\n"), 0644)
	require.NoError(t, err)

	tool := EditTool{}
	args, _ := json.Marshal(map[string]string{
		"path":    path,
		"old_str": "not present",
		"new_str": "replacement",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "old_str was not found in the file")
}

func TestEditTool_MultipleMatchesError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	err := os.WriteFile(path, []byte("foo bar foo baz foo\n"), 0644)
	require.NoError(t, err)

	tool := EditTool{}
	args, _ := json.Marshal(map[string]string{
		"path":    path,
		"old_str": "foo",
		"new_str": "qux",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "appears 3 times")
}

func TestEditTool_EmptyOldStr(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	err := os.WriteFile(path, []byte("hello\n"), 0644)
	require.NoError(t, err)

	tool := EditTool{}
	args, _ := json.Marshal(map[string]string{
		"path":    path,
		"old_str": "",
		"new_str": "replacement",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "old_str must not be empty")
}

func TestEditTool_IdenticalOldAndNew(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	err := os.WriteFile(path, []byte("hello world\n"), 0644)
	require.NoError(t, err)

	tool := EditTool{}
	args, _ := json.Marshal(map[string]string{
		"path":    path,
		"old_str": "hello",
		"new_str": "hello",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "identical")
}

func TestEditTool_MultilineReplacement(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	original := "line1\nline2\nline3\n"
	err := os.WriteFile(path, []byte(original), 0644)
	require.NoError(t, err)

	tool := EditTool{}
	args, _ := json.Marshal(map[string]string{
		"path":    path,
		"old_str": "line1\nline2",
		"new_str": "first\nsecond",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "first\nsecond\nline3\n", string(data))
}

func TestEditTool_ReplaceAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	err := os.WriteFile(path, []byte("foo bar foo baz foo\n"), 0644)
	require.NoError(t, err)

	tool := EditTool{}
	args, _ := json.Marshal(map[string]any{
		"path":        path,
		"old_str":     "foo",
		"new_str":     "qux",
		"replace_all": true,
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "qux bar qux baz qux\n", string(data))
}

func TestEditTool_MissingFile(t *testing.T) {
	tool := EditTool{}
	args, _ := json.Marshal(map[string]string{
		"path":    "/nonexistent/path/file.txt",
		"old_str": "something",
		"new_str": "else",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

func TestEditTool_InvalidJSON(t *testing.T) {
	tool := EditTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestEditTool_PreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "exec.sh")
	err := os.WriteFile(path, []byte("#!/bin/sh\necho hello\n"), 0755)
	require.NoError(t, err)

	tool := EditTool{}
	args, _ := json.Marshal(map[string]string{
		"path":    path,
		"old_str": "hello",
		"new_str": "world",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestEditTool_RelativePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rel.txt")
	err := os.WriteFile(path, []byte("hello world\n"), 0644)
	require.NoError(t, err)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()
	require.NoError(t, os.Chdir(dir))

	tool := EditTool{}
	args, _ := json.Marshal(map[string]string{
		"path":    "rel.txt",
		"old_str": "world",
		"new_str": "Go",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "hello Go\n", string(data))
}

// Verify the tool satisfies the interface
var _ tools.Tool = EditTool{}
