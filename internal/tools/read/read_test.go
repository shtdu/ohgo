package read

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

func TestReadTool_NameAndSchema(t *testing.T) {
	tool := ReadTool{}
	assert.Equal(t, "read_file", tool.Name())
	assert.Contains(t, tool.Description(), "file")
	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
	required := schema["required"].([]string)
	assert.Contains(t, required, "path")
}

func TestReadTool_FullFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	err := os.WriteFile(path, []byte("line1\nline2\nline3\n"), 0644)
	require.NoError(t, err)

	tool := ReadTool{}
	args, _ := json.Marshal(map[string]string{"path": path})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "1\tline1")
	assert.Contains(t, result.Content, "2\tline2")
	assert.Contains(t, result.Content, "3\tline3")
}

func TestReadTool_OffsetLimit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	var lines []string
	for i := 0; i < 10; i++ {
		lines = append(lines, "line"+strings.Repeat(" ", 10-i)+string(rune('0'+i)))
	}
	err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)
	require.NoError(t, err)

	tool := ReadTool{}
	args, _ := json.Marshal(map[string]any{"path": path, "offset": 2, "limit": 3})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "3\t")
	assert.Contains(t, result.Content, "5\t")
	assert.NotContains(t, result.Content, "6\t")
}

func TestReadTool_MissingFile(t *testing.T) {
	tool := ReadTool{}
	args, _ := json.Marshal(map[string]string{"path": "/nonexistent/file.txt"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

func TestReadTool_Directory(t *testing.T) {
	dir := t.TempDir()
	tool := ReadTool{}
	args, _ := json.Marshal(map[string]string{"path": dir})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "directory")
}

func TestReadTool_BinaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "binary.bin")
	binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00, 0x00} // PNG header with null
	err := os.WriteFile(path, binaryData, 0644)
	require.NoError(t, err)

	tool := ReadTool{}
	args, _ := json.Marshal(map[string]string{"path": path})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "Binary")
}

func TestReadTool_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	err := os.WriteFile(path, []byte(""), 0644)
	require.NoError(t, err)

	tool := ReadTool{}
	args, _ := json.Marshal(map[string]string{"path": path})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Empty(t, strings.TrimSpace(result.Content))
}

func TestReadTool_OffsetPastEnd(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "short.txt")
	err := os.WriteFile(path, []byte("only line\n"), 0644)
	require.NoError(t, err)

	tool := ReadTool{}
	args, _ := json.Marshal(map[string]any{"path": path, "offset": 100})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "no content in selected range")
}

func TestReadTool_RelativePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rel.txt")
	err := os.WriteFile(path, []byte("hello\n"), 0644)
	require.NoError(t, err)

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldWd) }()
	require.NoError(t, os.Chdir(dir))

	tool := ReadTool{}
	args, _ := json.Marshal(map[string]string{"path": "rel.txt"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "hello")
}

func TestReadTool_InvalidJSON(t *testing.T) {
	tool := ReadTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestReadTool_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tool := ReadTool{}
	args, _ := json.Marshal(map[string]string{"path": "go.mod"})
	_, err := tool.Execute(ctx, args)
	assert.Error(t, err)
}

// Verify the tool satisfies the interface
var _ tools.Tool = ReadTool{}
