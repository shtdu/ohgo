package todo

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

func TestTodoWriteTool_NameAndSchema(t *testing.T) {
	tool := TodoWriteTool{}
	assert.Equal(t, "todo_write", tool.Name())
	assert.Contains(t, tool.Description(), "task list")
	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
	required := schema["required"].([]string)
	assert.Contains(t, required, "todos")
}

func TestTodoWriteTool_WriteStructuredTodos(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")

	tool := TodoWriteTool{}
	args, _ := json.Marshal(map[string]any{
		"path": path,
		"todos": []map[string]string{
			{"subject": "Write tests", "status": "completed"},
			{"subject": "Implement feature", "description": "Add the new thing", "status": "in_progress"},
			{"subject": "Review PR", "status": "pending"},
		},
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "3 todo(s)")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)

	assert.Contains(t, content, "## Tasks")
	assert.Contains(t, content, "- [x] Write tests")
	assert.Contains(t, content, "- [~] Implement feature — Add the new thing")
	assert.Contains(t, content, "- [ ] Review PR")
}

func TestTodoWriteTool_OverwriteExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")

	// Write initial content
	err := os.WriteFile(path, []byte("old content"), 0644)
	require.NoError(t, err)

	tool := TodoWriteTool{}
	args, _ := json.Marshal(map[string]any{
		"path": path,
		"todos": []map[string]string{
			{"subject": "New task", "status": "pending"},
		},
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "- [ ] New task")
	assert.NotContains(t, string(data), "old content")
}

func TestTodoWriteTool_EmptyTodosClearsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TODO.md")

	// Write initial content
	err := os.WriteFile(path, []byte("existing tasks"), 0644)
	require.NoError(t, err)

	tool := TodoWriteTool{}
	args, _ := json.Marshal(map[string]any{
		"path":  path,
		"todos": []map[string]string{},
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "0 todo(s)")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(data), "## Tasks"))
	assert.NotContains(t, string(data), "existing tasks")
}

func TestTodoWriteTool_InvalidJSON(t *testing.T) {
	tool := TodoWriteTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestTodoWriteTool_InvalidStatus(t *testing.T) {
	tool := TodoWriteTool{}
	args, _ := json.Marshal(map[string]any{
		"todos": []map[string]string{
			{"subject": "Bad task", "status": "unknown"},
		},
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid status")
}

func TestTodoWriteTool_DefaultPath(t *testing.T) {
	dir := t.TempDir()
	// Change to temp dir so default TODO.md lands there
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	tool := TodoWriteTool{}
	args, _ := json.Marshal(map[string]any{
		"todos": []map[string]string{
			{"subject": "Default path task", "status": "pending"},
		},
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "TODO.md")

	data, err := os.ReadFile(filepath.Join(dir, "TODO.md"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "- [ ] Default path task")
}

func TestTodoWriteTool_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tool := TodoWriteTool{}
	args, _ := json.Marshal(map[string]any{
		"todos": []map[string]string{
			{"subject": "Task", "status": "pending"},
		},
	})
	_, err := tool.Execute(ctx, args)
	assert.Error(t, err)
}

// Verify the tool satisfies the interface
var _ tools.Tool = TodoWriteTool{}
