package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/api"
	"github.com/shtdu/ohgo/internal/permissions"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestExecuteTool_UnknownTool(t *testing.T) {
	registry := tools.NewRegistry()
	eng := New(Options{ToolReg: registry, MaxTurns: 10})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "nonexistent",
		Input: json.RawMessage(`{}`),
	})
	assert.True(t, isErr)
	assert.Contains(t, output, "unknown tool")
}

func TestExecuteTool_ToolReturnsError(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{
		name: "failing",
		execute: func(_ context.Context, _ json.RawMessage) (tools.Result, error) {
			return tools.Result{}, fmt.Errorf("something broke")
		},
	})
	eng := New(Options{ToolReg: registry, MaxTurns: 10})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "failing",
		Input: json.RawMessage(`{}`),
	})
	assert.True(t, isErr)
	assert.Contains(t, output, "tool error")
}

func TestExecuteTool_ToolReturnsIsError(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{
		name: "bad-input",
		execute: func(_ context.Context, _ json.RawMessage) (tools.Result, error) {
			return tools.Result{Content: "bad input", IsError: true}, nil
		},
	})
	eng := New(Options{ToolReg: registry, MaxTurns: 10})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "bad-input",
		Input: json.RawMessage(`{}`),
	})
	assert.True(t, isErr)
	assert.Equal(t, "bad input", output)
}

func TestExecuteTool_Success(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{
		name: "echo",
		execute: func(_ context.Context, args json.RawMessage) (tools.Result, error) {
			return tools.Result{Content: string(args)}, nil
		},
	})
	eng := New(Options{ToolReg: registry, MaxTurns: 10})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "echo",
		Input: json.RawMessage(`{"msg":"hi"}`),
	})
	assert.False(t, isErr)
	assert.Equal(t, `{"msg":"hi"}`, output)
}

// mockHookRunner implements hooks.HookRunner for testing.
type mockHookRunner struct {
	preBlocked bool
	preReason  string
	preErr     error
	postErr    error
}

func (m *mockHookRunner) RunPre(_ context.Context, _ string, _ map[string]any) (bool, string, error) {
	return m.preBlocked, m.preReason, m.preErr
}

func (m *mockHookRunner) RunPost(_ context.Context, _ string, _ map[string]any, _ any) error {
	return m.postErr
}

func TestExecuteTool_HookBlocks(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})

	eng := New(Options{
		ToolReg: registry,
		MaxTurns: 10,
		Hooks: &mockHookRunner{preBlocked: true, preReason: "not allowed"},
	})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "bash",
		Input: json.RawMessage(`{"command":"rm -rf /"}`),
	})
	assert.True(t, isErr)
	assert.Contains(t, output, "blocked by hook")
	assert.Contains(t, output, "not allowed")
}

func TestExecuteTool_HookError(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})

	eng := New(Options{
		ToolReg: registry,
		MaxTurns: 10,
		Hooks: &mockHookRunner{preErr: fmt.Errorf("hook crashed")},
	})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "bash",
		Input: json.RawMessage(`{"command":"ls"}`),
	})
	assert.True(t, isErr)
	assert.Contains(t, output, "hook error")
}

func TestExecuteTool_PostHook(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "read_file"})

	eng := New(Options{
		ToolReg: registry,
		MaxTurns: 10,
		Hooks: &mockHookRunner{}, // noop hooks
	})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "read_file",
		Input: json.RawMessage(`{"file_path":"/tmp/test"}`),
	})
	assert.False(t, isErr)
	assert.Equal(t, "mock result", output)
}

// mockChecker implements permissions.Checker for testing.
type mockChecker struct {
	decision permissions.Decision
	err      error
}

func (m *mockChecker) Check(_ context.Context, _ permissions.Check) (permissions.Decision, error) {
	return m.decision, m.err
}

func TestExecuteTool_PermissionDenied(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})

	eng := New(Options{
		ToolReg:    registry,
		MaxTurns:   10,
		Permission: &mockChecker{decision: permissions.Deny},
	})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "bash",
		Input: json.RawMessage(`{"command":"ls"}`),
	})
	assert.True(t, isErr)
	assert.Contains(t, output, "denied by permissions")
}

func TestExecuteTool_PermissionAsk_UserDenies(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})

	eng := New(Options{
		ToolReg:    registry,
		MaxTurns:   10,
		Permission: &mockChecker{decision: permissions.Ask},
		PermPrompt: stubPrompter{allow: false},
	})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "bash",
		Input: json.RawMessage(`{"command":"ls"}`),
	})
	assert.True(t, isErr)
	assert.Contains(t, output, "denied by user")
}

func TestExecuteTool_PermissionAsk_UserAllows(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})

	eng := New(Options{
		ToolReg:    registry,
		MaxTurns:   10,
		Permission: &mockChecker{decision: permissions.Ask},
		PermPrompt: stubPrompter{allow: true},
	})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "bash",
		Input: json.RawMessage(`{"command":"ls"}`),
	})
	assert.False(t, isErr)
	assert.Equal(t, "mock result", output)
}

func TestExecuteTool_PermissionAsk_NoPrompt(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})

	eng := New(Options{
		ToolReg:    registry,
		MaxTurns:   10,
		Permission: &mockChecker{decision: permissions.Ask},
		// PermPrompt is nil
	})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "bash",
		Input: json.RawMessage(`{"command":"ls"}`),
	})
	assert.True(t, isErr)
	assert.Contains(t, output, "non-interactive mode")
}

func TestExecuteTool_PermissionCheckError(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})

	eng := New(Options{
		ToolReg:    registry,
		MaxTurns:   10,
		Permission: &mockChecker{err: fmt.Errorf("checker down")},
	})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "bash",
		Input: json.RawMessage(`{"command":"ls"}`),
	})
	assert.True(t, isErr)
	assert.Contains(t, output, "permission check error")
}

func TestExecuteTool_PermPromptError(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})

	eng := New(Options{
		ToolReg:    registry,
		MaxTurns:   10,
		Permission: &mockChecker{decision: permissions.Ask},
		PermPrompt: stubPrompter{err: fmt.Errorf("prompt crashed")},
	})

	output, isErr := eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "bash",
		Input: json.RawMessage(`{"command":"ls"}`),
	})
	assert.True(t, isErr)
	assert.Contains(t, output, "permission prompt error")
}

func TestExecuteTool_ExtractsFilePath(t *testing.T) {
	var capturedCheck permissions.Check
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "read_file"})

	eng := New(Options{
		ToolReg:  registry,
		MaxTurns: 10,
		Permission: &captureChecker{capture: &capturedCheck},
	})

	_, _ = eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "read_file",
		Input: json.RawMessage(`{"file_path":"/tmp/test.txt"}`),
	})
	assert.Equal(t, "/tmp/test.txt", capturedCheck.FilePath)
	assert.Equal(t, "read_file", capturedCheck.ToolName)
}

func TestExecuteTool_ExtractsCommand(t *testing.T) {
	var capturedCheck permissions.Check
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})

	eng := New(Options{
		ToolReg:  registry,
		MaxTurns: 10,
		Permission: &captureChecker{capture: &capturedCheck},
	})

	_, _ = eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "bash",
		Input: json.RawMessage(`{"command":"ls -la"}`),
	})
	assert.Equal(t, "ls -la", capturedCheck.Command)
}

func TestExecuteTool_PathField(t *testing.T) {
	var capturedCheck permissions.Check
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "glob"})

	eng := New(Options{
		ToolReg:  registry,
		MaxTurns: 10,
		Permission: &captureChecker{capture: &capturedCheck},
	})

	_, _ = eng.executeTool(context.Background(), api.ToolCall{
		ID:   "t1",
		Name: "glob",
		Input: json.RawMessage(`{"path":"/src/**/*.go"}`),
	})
	assert.Equal(t, "/src/**/*.go", capturedCheck.FilePath)
}

func TestBuildToolDefs(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})
	registry.Register(&mockTool{name: "read_file"})

	eng := New(Options{ToolReg: registry, MaxTurns: 10})
	defs := eng.buildToolDefs()
	require.Len(t, defs, 2)

	names := map[string]bool{}
	for _, d := range defs {
		names[d.Name] = true
	}
	assert.True(t, names["bash"])
	assert.True(t, names["read_file"])
}

func TestBuildToolDefs_NilRegistry(t *testing.T) {
	eng := New(Options{MaxTurns: 10})
	defs := eng.buildToolDefs()
	assert.Nil(t, defs)
}

func TestEmit_ChannelFull(t *testing.T) {
	ch := make(chan EngineEvent, 1)
	ch <- EngineEvent{Type: EventTextDelta} // fill channel

	eng := New(Options{EventCh: ch, MaxTurns: 10})
	// Should not block when channel is full
	eng.emit(EngineEvent{Type: EventTextDelta, Data: "overflow"})
}

func TestEmit_NilChannel(t *testing.T) {
	eng := New(Options{MaxTurns: 10})
	// Should not panic with nil channel
	eng.emit(EngineEvent{Type: EventTextDelta, Data: "test"})
}

// stubPrompter implements PermissionPrompter for testing.
type stubPrompter struct {
	allow    bool
	remember bool
	err      error
}

func (s stubPrompter) PromptApproval(_ context.Context, _ string, _ string) (bool, bool, error) {
	return s.allow, s.remember, s.err
}

// captureChecker records the check and allows.
type captureChecker struct {
	capture *permissions.Check
}

func (c *captureChecker) Check(_ context.Context, check permissions.Check) (permissions.Decision, error) {
	*c.capture = check
	return permissions.Allow, nil
}
