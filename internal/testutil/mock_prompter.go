//go:build integration

package testutil

import (
	"context"
	"sync"
)

// PromptResponse configures what a MockPermissionPrompter returns for a given tool.
type PromptResponse struct {
	Allow    bool
	Remember bool
	Error    error
}

// MockPermissionPrompter implements engine.PermissionPrompter for testing.
// It records all calls and returns configurable responses per tool name.
type MockPermissionPrompter struct {
	mu       sync.Mutex
	responses map[string]PromptResponse
	default_  PromptResponse
	calls    []PromptCall
}

// PromptCall records a single prompt invocation.
type PromptCall struct {
	ToolName string
	Details  string
}

// NewMockPermissionPrompter creates a prompter with the given responses.
// toolResponses maps tool name to response; unmapped tools use default_.
func NewMockPermissionPrompter(defaultResponse PromptResponse, toolResponses map[string]PromptResponse) *MockPermissionPrompter {
	if toolResponses == nil {
		toolResponses = make(map[string]PromptResponse)
	}
	return &MockPermissionPrompter{
		responses: toolResponses,
		default_:  defaultResponse,
	}
}

// PromptApproval implements engine.PermissionPrompter.
func (m *MockPermissionPrompter) PromptApproval(_ context.Context, toolName string, details string) (bool, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, PromptCall{ToolName: toolName, Details: details})

	resp, ok := m.responses[toolName]
	if !ok {
		resp = m.default_
	}
	return resp.Allow, resp.Remember, resp.Error
}

// Calls returns a copy of recorded prompt calls.
func (m *MockPermissionPrompter) Calls() []PromptCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]PromptCall, len(m.calls))
	copy(out, m.calls)
	return out
}

// CallCount returns the number of prompt calls made.
func (m *MockPermissionPrompter) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

// Responses returns the response map for per-tool configuration.
func (m *MockPermissionPrompter) Responses() map[string]PromptResponse {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.responses
}
