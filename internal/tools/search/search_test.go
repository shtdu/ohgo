package search

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

type mockTool struct {
	name string
	desc string
}

func (m mockTool) Name() string                                            { return m.name }
func (m mockTool) Description() string                                     { return m.desc }
func (m mockTool) InputSchema() map[string]any                             { return map[string]any{"type": "object"} }
func (m mockTool) Execute(_ context.Context, _ json.RawMessage) (tools.Result, error) {
	return tools.Result{}, nil
}

func newRegistryWithTools(toolList ...mockTool) *tools.Registry {
	r := tools.NewRegistry()
	for _, t := range toolList {
		r.Register(t)
	}
	return r
}

func TestSearchTool_Name(t *testing.T) {
	assert.Equal(t, "tool_search", SearchTool{}.Name())
}

func TestSearchTool_MatchByName(t *testing.T) {
	registry := newRegistryWithTools(
		mockTool{name: "bash", desc: "Run shell commands"},
		mockTool{name: "read_file", desc: "Read a file"},
		mockTool{name: "grep", desc: "Search file contents"},
	)

	tool := SearchTool{Registry: registry}
	args, _ := json.Marshal(map[string]string{"query": "bash"})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "bash: Run shell commands")
	assert.NotContains(t, result.Content, "read_file")
}

func TestSearchTool_MatchByDescription(t *testing.T) {
	registry := newRegistryWithTools(
		mockTool{name: "bash", desc: "Run shell commands"},
		mockTool{name: "read_file", desc: "Read a file"},
		mockTool{name: "grep", desc: "Search file contents"},
	)

	tool := SearchTool{Registry: registry}
	args, _ := json.Marshal(map[string]string{"query": "file"})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "read_file: Read a file")
	assert.Contains(t, result.Content, "grep: Search file contents")
}

func TestSearchTool_NoMatches(t *testing.T) {
	registry := newRegistryWithTools(
		mockTool{name: "bash", desc: "Run shell commands"},
	)

	tool := SearchTool{Registry: registry}
	args, _ := json.Marshal(map[string]string{"query": "nonexistent"})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "(no matching tools found)", result.Content)
}

func TestSearchTool_EmptyQuery(t *testing.T) {
	registry := newRegistryWithTools(
		mockTool{name: "bash", desc: "Run shell commands"},
	)

	tool := SearchTool{Registry: registry}
	args, _ := json.Marshal(map[string]string{"query": ""})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Equal(t, "query is required", result.Content)
}

func TestSearchTool_CaseInsensitive(t *testing.T) {
	registry := newRegistryWithTools(
		mockTool{name: "Bash", desc: "Run Shell Commands"},
		mockTool{name: "grep", desc: "Search file contents"},
	)

	tool := SearchTool{Registry: registry}
	args, _ := json.Marshal(map[string]string{"query": "BASH"})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Bash: Run Shell Commands")

	// Also test lowercase query against mixed-case tool
	args2, _ := json.Marshal(map[string]string{"query": "shell"})
	result2, err := tool.Execute(context.Background(), args2)
	require.NoError(t, err)
	assert.False(t, result2.IsError)
	assert.Contains(t, result2.Content, "Bash: Run Shell Commands")
}

func TestSearchTool_InvalidJSON(t *testing.T) {
	tool := SearchTool{Registry: tools.NewRegistry()}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestSearchTool_NilRegistry(t *testing.T) {
	tool := SearchTool{Registry: nil}
	args, _ := json.Marshal(map[string]string{"query": "test"})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Equal(t, "tool registry not available", result.Content)
}

func TestSearchTool_MultipleResults(t *testing.T) {
	registry := newRegistryWithTools(
		mockTool{name: "read_file", desc: "Read a file from disk"},
		mockTool{name: "write_file", desc: "Write content to a file"},
		mockTool{name: "bash", desc: "Run shell commands"},
	)

	tool := SearchTool{Registry: registry}
	args, _ := json.Marshal(map[string]string{"query": "file"})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	lines := strings.Split(result.Content, "\n")
	assert.Equal(t, 2, len(lines))
	assert.Contains(t, result.Content, "read_file")
	assert.Contains(t, result.Content, "write_file")
}

var _ tools.Tool = SearchTool{}
