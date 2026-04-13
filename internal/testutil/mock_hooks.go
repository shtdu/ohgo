//go:build integration

package testutil

import (
	"context"
	"sync"
)

// MockHookRunner implements hooks.HookRunner for testing.
// It records all hook invocations and supports configurable blocking.
type MockHookRunner struct {
	mu         sync.Mutex
	preBlocked bool
	preReason  string
	preErr     error
	postErr    error

	preCalls  []HookCall
	postCalls []HookCall
}

// HookCall records a single hook invocation.
type HookCall struct {
	ToolName string
	Args     map[string]any
}

// NewMockHookRunner creates a hook runner with the given pre/post behavior.
func NewMockHookRunner(preBlocked bool, preReason string) *MockHookRunner {
	return &MockHookRunner{preBlocked: preBlocked, preReason: preReason}
}

// RunPre implements hooks.HookRunner.
func (m *MockHookRunner) RunPre(_ context.Context, toolName string, args map[string]any) (bool, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.preCalls = append(m.preCalls, HookCall{ToolName: toolName, Args: args})
	return m.preBlocked, m.preReason, m.preErr
}

// RunPost implements hooks.HookRunner.
func (m *MockHookRunner) RunPost(_ context.Context, toolName string, args map[string]any, result any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.postCalls = append(m.postCalls, HookCall{ToolName: toolName, Args: args})
	return m.postErr
}

// PreCalls returns recorded pre-hook calls.
func (m *MockHookRunner) PreCalls() []HookCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]HookCall, len(m.preCalls))
	copy(out, m.preCalls)
	return out
}

// PostCalls returns recorded post-hook calls.
func (m *MockHookRunner) PostCalls() []HookCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]HookCall, len(m.postCalls))
	copy(out, m.postCalls)
	return out
}
