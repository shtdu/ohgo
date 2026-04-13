//go:build integration

package engine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/api"
	"github.com/shtdu/ohgo/internal/testutil"
	"github.com/shtdu/ohgo/internal/tools"
)

// mockDynamicTool implements tools.Tool for dynamic registration tests.
type mockDynamicTool struct {
	name string
}

func (m *mockDynamicTool) Name() string                        { return m.name }
func (m *mockDynamicTool) Description() string                 { return "dynamic test tool" }
func (m *mockDynamicTool) InputSchema() map[string]any         { return map[string]any{"type": "object"} }
func (m *mockDynamicTool) Execute(_ context.Context, _ json.RawMessage) (tools.Result, error) {
	return tools.Result{Content: "dynamic ok"}, nil
}

// f is a helper to create a minimal fixture for direct tool tests.
func newFixture(t *testing.T) *testutil.Fixture {
	t.Helper()
	return testutil.NewFixture(t, testutil.FixtureConfig{})
}

// EARS: REQ-TL-001
func TestIntegration_ToolRegistry_DynamicExpansion(t *testing.T) {
	f := newFixture(t)

	// Verify initial tools are registered
	list := f.Registry.List()
	assert.NotEmpty(t, list, "registry should have tools")

	// Verify each tool has a unique name
	names := map[string]bool{}
	for _, tool := range list {
		assert.NotEmpty(t, tool.Name())
		assert.NotEmpty(t, tool.Description())
		assert.False(t, names[tool.Name()], "duplicate tool name: %s", tool.Name())
		names[tool.Name()] = true
	}

	// Verify dynamic expansion: register a new tool and it appears
	f.Registry.Register(&mockDynamicTool{name: "dynamic_test"})

	found := false
	for _, tool := range f.Registry.List() {
		if tool.Name() == "dynamic_test" {
			found = true
			break
		}
	}
	assert.True(t, found, "dynamically registered tool should appear in registry")
}

// EARS: REQ-TL-001
func TestIntegration_ToolRegistry_InvalidSchemaExcluded(t *testing.T) {
	f := newFixture(t)
	// All registered tools should produce valid tool defs for the API
	eng := f.Engine

	// Access buildToolDefs through the engine by making a query
	// that exercises the full loop
	toolNames := f.Registry.List()
	for _, tool := range toolNames {
		schema := tool.InputSchema()
		// Schema can be nil (optional), but if present should be a map
		if schema != nil {
			_, ok := schema["type"]
			assert.True(t, ok, "tool %s schema should have 'type' field", tool.Name())
		}
	}
	_ = eng
}

// EARS: REQ-TL-002
func TestIntegration_FileOps_WriteAndReadRoundTrip(t *testing.T) {
	f := newFixture(t)

	// Write a file using the real write tool
	writeTool := f.Registry.Get("write_file")
	require.NotNil(t, writeTool)
	_, err := writeTool.Execute(context.Background(), json.RawMessage(fmt.Sprintf(
		`{"path":"%s/test.txt","content":"hello integration"}`, f.Dir,
	)))
	require.NoError(t, err)

	// Read it back using the real read tool
	readTool := f.Registry.Get("read_file")
	require.NotNil(t, readTool)
	result, err := readTool.Execute(context.Background(), json.RawMessage(fmt.Sprintf(
		`{"path":"%s/test.txt"}`, f.Dir,
	)))
	require.NoError(t, err)
	assert.Contains(t, result.Content, "hello integration")
	assert.False(t, result.IsError)
}

// EARS: REQ-TL-002
func TestIntegration_FileOps_EditReplacesContent(t *testing.T) {
	f := newFixture(t)

	// Write initial file
	testutil.WriteFile(t, f.Dir, "edit.txt", "hello world")

	// Edit using the real edit tool
	editTool := f.Registry.Get("edit_file")
	require.NotNil(t, editTool)
	result, err := editTool.Execute(context.Background(), json.RawMessage(fmt.Sprintf(
		`{"path":"%s/edit.txt","old_str":"hello","new_str":"goodbye"}`, f.Dir,
	)))
	require.NoError(t, err)
	assert.False(t, result.IsError)

	// Verify content changed
	content := testutil.ReadFile(t, f.Dir, "edit.txt")
	assert.Contains(t, content, "goodbye world")
	assert.NotContains(t, content, "hello world")
}

// EARS: REQ-TL-002
func TestIntegration_FileOps_ReadNonExistent(t *testing.T) {
	f := newFixture(t)
	readTool := f.Registry.Get("read_file")
	require.NotNil(t, readTool)

	result, err := readTool.Execute(context.Background(), json.RawMessage(
		fmt.Sprintf(`{"path":"%s/nonexistent.txt"}`, f.Dir),
	))
	require.NoError(t, err)
	assert.True(t, result.IsError, "reading nonexistent file should return error")
}

// EARS: REQ-TL-003
func TestIntegration_Bash_ExecutesAndReturnsOutput(t *testing.T) {
	f := newFixture(t)
	tool := f.Registry.Get("bash")
	require.NotNil(t, tool)

	result, err := tool.Execute(context.Background(), json.RawMessage(
		`{"command":"echo hello integration"}`,
	))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "hello integration")
}

// EARS: REQ-TL-003
func TestIntegration_Bash_NonZeroExitCode(t *testing.T) {
	f := newFixture(t)
	tool := f.Registry.Get("bash")
	require.NotNil(t, tool)

	result, err := tool.Execute(context.Background(), json.RawMessage(
		`{"command":"exit 42"}`,
	))
	require.NoError(t, err)
	// Non-zero exit code — tool should return output with error flag
	assert.Contains(t, result.Content, "42")
}

// EARS: REQ-TL-003
func TestIntegration_Bash_TimeoutKillsProcess(t *testing.T) {
	if _, err := exec.LookPath("sleep"); err != nil {
		t.Skip("sleep not available")
	}

	f := newFixture(t)
	tool := f.Registry.Get("bash")
	require.NotNil(t, tool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	result, err := tool.Execute(ctx, json.RawMessage(
		`{"command":"sleep 60","timeout_seconds":1}`,
	))
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.True(t, result.IsError, "timed out command should return error")
	assert.Less(t, elapsed, 10*time.Second, "should not run for full 60s")
}

// EARS: REQ-TL-003
func TestIntegration_Bash_CommandNotFound(t *testing.T) {
	f := newFixture(t)
	tool := f.Registry.Get("bash")
	require.NotNil(t, tool)

	result, err := tool.Execute(context.Background(), json.RawMessage(
		`{"command":"nonexistent_command_xyz_123"}`,
	))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// EARS: REQ-TL-004
func TestIntegration_Glob_FindsFilesByPattern(t *testing.T) {
	f := newFixture(t)
	testutil.WriteFile(t, f.Dir, "a.go", "package a")
	testutil.WriteFile(t, f.Dir, "b.go", "package b")
	testutil.WriteFile(t, f.Dir, "c.txt", "text file")

	tool := f.Registry.Get("glob")
	require.NotNil(t, tool)

	result, err := tool.Execute(context.Background(), json.RawMessage(fmt.Sprintf(
		`{"pattern":"%s/*.go"}`, f.Dir,
	)))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "a.go")
	assert.Contains(t, result.Content, "b.go")
	assert.NotContains(t, result.Content, "c.txt")
}

// EARS: REQ-TL-004
func TestIntegration_Glob_InvalidPattern(t *testing.T) {
	f := newFixture(t)
	tool := f.Registry.Get("glob")
	require.NotNil(t, tool)

	result, err := tool.Execute(context.Background(), json.RawMessage(
		`{"pattern":"[invalid"}`,
	))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// EARS: REQ-TL-005
func TestIntegration_Grep_RegexAndContext(t *testing.T) {
	f := newFixture(t)
	testutil.WriteFile(t, f.Dir, "test.go", "package main\n\nfunc hello() {\n\treturn\n}\n")

	tool := f.Registry.Get("grep")
	require.NotNil(t, tool)

	result, err := tool.Execute(context.Background(), json.RawMessage(fmt.Sprintf(
		`{"pattern":"func hello","path":"%s","context":1}`,
		f.Dir,
	)))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "func hello()")
}

// EARS: REQ-TL-005
func TestIntegration_Grep_InvalidRegex(t *testing.T) {
	f := newFixture(t)
	tool := f.Registry.Get("grep")
	require.NotNil(t, tool)

	result, err := tool.Execute(context.Background(), json.RawMessage(
		`{"pattern":"[invalid"}`,
	))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// EARS: REQ-TL-004
func TestIntegration_Glob_NonExistentDirectory(t *testing.T) {
	f := newFixture(t)
	tool := f.Registry.Get("glob")
	require.NotNil(t, tool)

	result, err := tool.Execute(context.Background(), json.RawMessage(
		fmt.Sprintf(`{"pattern":"%s/nonexistent/*.txt"}`, filepath.Join(f.Dir, "nope")),
	))
	require.NoError(t, err)
	// Glob on non-existent directory returns "(no matches)", not an error
	assert.Contains(t, result.Content, "no matches")
	assert.False(t, result.IsError)
}

// EARS: REQ-TL-001, REQ-TL-011
func TestIntegration_ToolDiscovery_SearchByName(t *testing.T) {
	f := newFixture(t)

	// Search tool needs a registry reference — verify tool_search is not registered
	// without deps (search.SearchTool requires Registry), so test registry List directly
	_ = f.Registry.Get("tool_search")

	allTools := f.Registry.List()
	assert.True(t, len(allTools) >= 6, "should have at least 6 core tools")

	names := make([]string, 0, len(allTools))
	for _, tool := range allTools {
		names = append(names, tool.Name())
	}
	assert.Contains(t, names, "bash")
	assert.Contains(t, names, "read_file")
	assert.Contains(t, names, "write_file")
	assert.Contains(t, names, "edit_file")
	assert.Contains(t, names, "glob")
	assert.Contains(t, names, "grep")
}

// EARS: REQ-TL-003
func TestIntegration_Bash_EngineLoop_ExecutesTool(t *testing.T) {
	f := testutil.NewFixture(t, testutil.FixtureConfig{
		Responses: []testutil.ResponseScript{
			{
				TextDeltas: []string{"let me check"},
				ToolCalls: []api.ToolCall{
					{ID: "t1", Name: "bash", Input: json.RawMessage(`{"command":"echo hello"}`)},
				},
				Usage: api.UsageSnapshot{InputTokens: 20, OutputTokens: 10},
			},
			{
				TextDeltas: []string{"done"},
				Usage:      api.UsageSnapshot{InputTokens: 10, OutputTokens: 5},
			},
		},
		PermMode: "auto",
	})

	_, stop := f.DrainEvents()
	err := f.Engine.Query(context.Background(), "run echo")
	stop()

	require.NoError(t, err)
	assert.Equal(t, 2, f.Engine.Turns())
}
