//go:build integration

package testutil

import (
	"context"
	"sync"

	"github.com/shtdu/ohgo/internal/api"
)

// ResponseScript defines one API response in a multi-turn test.
type ResponseScript struct {
	TextDeltas []string
	ToolCalls  []api.ToolCall
	Usage      api.UsageSnapshot
	// Error, if non-nil, causes Stream to return this error.
	Error error
}

// MockAPIClient implements api.Client with scripted responses.
// It records all Stream calls for post-test assertions.
type MockAPIClient struct {
	mu        sync.Mutex
	responses []ResponseScript
	callIndex int
	calls     []api.StreamOptions // recorded StreamOptions per call
}

// NewMockAPIClient creates a mock client with the given response script.
func NewMockAPIClient(responses ...ResponseScript) *MockAPIClient {
	return &MockAPIClient{responses: responses}
}

// Stream implements api.Client. Each call advances to the next scripted response.
func (m *MockAPIClient) Stream(ctx context.Context, opts api.StreamOptions) (<-chan api.StreamEvent, error) {
	m.mu.Lock()
	idx := m.callIndex
	m.callIndex++
	m.calls = append(m.calls, opts)
	m.mu.Unlock()

	ch := make(chan api.StreamEvent, 64)

	if idx >= len(m.responses) {
		go func() {
			defer close(ch)
			ch <- api.StreamEvent{Type: "error", Data: "no more scripted responses"}
		}()
		return ch, nil
	}

	resp := m.responses[idx]
	if resp.Error != nil {
		return nil, resp.Error
	}

	go func() {
		defer close(ch)

		// Emit text deltas
		for _, text := range resp.TextDeltas {
			ch <- api.StreamEvent{Type: "text_delta", Data: text}
		}

		// Build content blocks for message_complete
		var blocks []api.ContentBlock
		for _, text := range resp.TextDeltas {
			blocks = append(blocks, api.ContentBlock{Type: "text", Text: text})
		}
		for _, tc := range resp.ToolCalls {
			blocks = append(blocks, api.ContentBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Name,
				Input: tc.Input,
			})
		}

		ch <- api.StreamEvent{Type: "message_complete", Data: api.NewAssistantMessage(blocks)}
		ch <- api.StreamEvent{Type: "usage", Data: resp.Usage}
	}()

	return ch, nil
}

// CallCount returns the number of Stream calls made.
func (m *MockAPIClient) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callIndex
}

// Calls returns a copy of the recorded StreamOptions.
func (m *MockAPIClient) Calls() []api.StreamOptions {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]api.StreamOptions, len(m.calls))
	copy(out, m.calls)
	return out
}
