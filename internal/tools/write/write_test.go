package write

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestWriteTool_NameAndSchema(t *testing.T) {
	tool := WriteTool{}
	assert.Equal(t, "write_file", tool.Name())
	assert.Contains(t, tool.Description(), "file")
	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
	required := schema["required"].([]string)
	assert.Contains(t, required, "path")
	assert.Contains(t, required, "content")
}

func TestWriteTool_CreateNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.txt")

	tool := WriteTool{}
	args, _ := json.Marshal(map[string]string{"path": path, "content": "hello world"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Wrote")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(data))
}

func TestWriteTool_OverwriteExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.txt")
	err := os.WriteFile(path, []byte("old content"), 0644)
	require.NoError(t, err)

	tool := WriteTool{}
	args, _ := json.Marshal(map[string]string{"path": path, "content": "new content"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Wrote")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "new content", string(data))
}

func TestWriteTool_CreateParentDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "file.txt")

	tool := WriteTool{}
	args, _ := json.Marshal(map[string]string{"path": path, "content": "nested"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Wrote")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "nested", string(data))
}

func TestWriteTool_WriteEmptyContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")

	tool := WriteTool{}
	args, _ := json.Marshal(map[string]string{"path": path, "content": ""})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Wrote")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "", string(data))
}

func TestWriteTool_ReadOnlyDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits not enforced on Windows")
	}

	dir := t.TempDir()
	readOnlyDir := filepath.Join(dir, "readonly")
	err := os.MkdirAll(readOnlyDir, 0555)
	require.NoError(t, err)
	// Ensure we can clean up after the test
	defer os.Chmod(readOnlyDir, 0755)

	path := filepath.Join(readOnlyDir, "file.txt")

	tool := WriteTool{}
	args, _ := json.Marshal(map[string]string{"path": path, "content": "should fail"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "Cannot write file")
}

func TestWriteTool_WriteToDirectory(t *testing.T) {
	dir := t.TempDir()

	tool := WriteTool{}
	args, _ := json.Marshal(map[string]string{"path": dir, "content": "data"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "directory")
}

func TestWriteTool_InvalidJSON(t *testing.T) {
	tool := WriteTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestWriteTool_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dir := t.TempDir()
	path := filepath.Join(dir, "cancelled.txt")

	tool := WriteTool{}
	args, _ := json.Marshal(map[string]string{"path": path, "content": "data"})
	_, err := tool.Execute(ctx, args)
	assert.Error(t, err)
}

// Verify the tool satisfies the interface
var _ tools.Tool = WriteTool{}
