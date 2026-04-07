package engine

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/api"
	"github.com/shtdu/ohgo/internal/tools"
)

// mockAPIClient implements api.Client for testing.
type mockAPIClient struct {
	responses []mockResponse
	callIndex int
}

type mockResponse struct {
	textDeltas []string
	toolCalls  []api.ToolCall
	usage      api.UsageSnapshot
}

func (m *mockAPIClient) Stream(ctx context.Context, opts api.StreamOptions) (<-chan api.StreamEvent, error) {
	ch := make(chan api.StreamEvent, 64)

	go func() {
		defer close(ch)
		if m.callIndex >= len(m.responses) {
			ch <- api.StreamEvent{Type: "error", Data: "no more responses"}
			return
		}
		resp := m.responses[m.callIndex]
		m.callIndex++

		// Emit text deltas
		for _, text := range resp.textDeltas {
			ch <- api.StreamEvent{Type: "text_delta", Data: text}
		}

		// Build content blocks
		var blocks []api.ContentBlock
		for _, text := range resp.textDeltas {
			blocks = append(blocks, api.ContentBlock{Type: "text", Text: text})
		}
		for _, tc := range resp.toolCalls {
			blocks = append(blocks, api.ContentBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Name,
				Input: tc.Input,
			})
		}

		// Emit message complete
		ch <- api.StreamEvent{Type: "message_complete", Data: api.NewAssistantMessage(blocks)}
		ch <- api.StreamEvent{Type: "usage", Data: resp.usage}
	}()

	return ch, nil
}

// mockTool implements tools.Tool for testing.
type mockTool struct {
	name    string
	execute func(ctx context.Context, args json.RawMessage) (tools.Result, error)
}

func (m *mockTool) Name() string                          { return m.name }
func (m *mockTool) Description() string                   { return "mock tool" }
func (m *mockTool) InputSchema() map[string]any           { return nil }
func (m *mockTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	if m.execute != nil {
		return m.execute(ctx, args)
	}
	return tools.Result{Content: "mock result"}, nil
}

func TestEngine_SimpleTextResponse(t *testing.T) {
	mockAPI := &mockAPIClient{
		responses: []mockResponse{
			{
				textDeltas: []string{"hello"},
				usage:      api.UsageSnapshot{InputTokens: 10, OutputTokens: 5},
			},
		},
	}

	eventCh := make(chan EngineEvent, 64)
	eng := New(Options{
		Model:     "test-model",
		MaxTokens: 100,
		MaxTurns:  10,
		APIClient: mockAPI,
		EventCh:   eventCh,
	})

	go func() {
		for range eventCh {
		}
	}()

	err := eng.Query(context.Background(), "hi")
	require.NoError(t, err)
	assert.Equal(t, 1, eng.costTracker.Turns())
	assert.Equal(t, 10, eng.TotalUsage().InputTokens)
}

func TestEngine_ToolUseLoop(t *testing.T) {
	mockAPI := &mockAPIClient{
		responses: []mockResponse{
			{
				textDeltas: []string{"let me check"},
				toolCalls: []api.ToolCall{
					{ID: "t1", Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)},
				},
				usage: api.UsageSnapshot{InputTokens: 20, OutputTokens: 10},
			},
			{
				textDeltas: []string{"here are the files"},
				usage:      api.UsageSnapshot{InputTokens: 30, OutputTokens: 15},
			},
		},
	}

	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})

	eventCh := make(chan EngineEvent, 64)
	eng := New(Options{
		Model:     "test-model",
		MaxTokens: 100,
		MaxTurns:  10,
		APIClient: mockAPI,
		ToolReg:   registry,
		EventCh:   eventCh,
	})

	var events []EngineEvent
	done := make(chan struct{})
	go func() {
		for e := range eventCh {
			events = append(events, e)
		}
		close(done)
	}()

	err := eng.Query(context.Background(), "list files")
	require.NoError(t, err)

	close(eventCh)
	<-done

	// Should have 2 turns (tool_use + text response)
	assert.Equal(t, 2, eng.costTracker.Turns())

	// Check events: text delta, turn complete, tool started, tool completed, text delta, turn complete
	var textDeltas, toolStarts, toolCompletes, turnCompletes int
	for _, e := range events {
		switch e.Type {
		case EventTextDelta:
			textDeltas++
		case EventToolStarted:
			toolStarts++
		case EventToolCompleted:
			toolCompletes++
		case EventTurnComplete:
			turnCompletes++
		}
	}
	assert.Equal(t, 2, textDeltas)
	assert.Equal(t, 1, toolStarts)
	assert.Equal(t, 1, toolCompletes)
	assert.Equal(t, 2, turnCompletes)
}

func TestEngine_MaxTurnsExceeded(t *testing.T) {
	// API always returns a tool_use
	mockAPI := &mockAPIClient{
		responses: []mockResponse{
			{toolCalls: []api.ToolCall{{ID: "t1", Name: "bash"}}, usage: api.UsageSnapshot{}},
			{toolCalls: []api.ToolCall{{ID: "t2", Name: "bash"}}, usage: api.UsageSnapshot{}},
			{toolCalls: []api.ToolCall{{ID: "t3", Name: "bash"}}, usage: api.UsageSnapshot{}},
		},
	}

	registry := tools.NewRegistry()
	registry.Register(&mockTool{name: "bash"})

	eng := New(Options{
		MaxTurns:  2,
		APIClient: mockAPI,
		ToolReg:   registry,
	})

	err := eng.Query(context.Background(), "do stuff")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max turns")
}

func TestEngine_ContextCancel(t *testing.T) {
	slowAPI := &slowAPIClient{}
	ctx, cancel := context.WithCancel(context.Background())

	eng := New(Options{
		MaxTurns:  10,
		APIClient: slowAPI,
	})

	// Cancel context before query returns
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// Query should complete (the slow client returns empty stream on cancel)
	// No assertion on error — the test verifies no deadlock/hang
	eng.Query(ctx, "hello")
}

type slowAPIClient struct{}

func (s *slowAPIClient) Stream(ctx context.Context, opts api.StreamOptions) (<-chan api.StreamEvent, error) {
	ch := make(chan api.StreamEvent)
	go func() {
		defer close(ch)
		<-ctx.Done()
	}()
	return ch, nil
}

func TestEngine_UnknownTool(t *testing.T) {
	mockAPI := &mockAPIClient{
		responses: []mockResponse{
			{
				toolCalls: []api.ToolCall{{ID: "t1", Name: "nonexistent"}},
				usage:     api.UsageSnapshot{},
			},
			{textDeltas: []string{"ok"}, usage: api.UsageSnapshot{}},
		},
	}

	registry := tools.NewRegistry()

	eng := New(Options{
		MaxTurns:  10,
		APIClient: mockAPI,
		ToolReg:   registry,
	})

	err := eng.Query(context.Background(), "test")
	require.NoError(t, err)
	// Should complete with error in tool result but not crash
}

func TestEngine_Clear(t *testing.T) {
	eng := New(Options{MaxTurns: 10})
	eng.messages = []api.Message{api.NewUserTextMessage("test")}
	eng.costTracker.Add(api.UsageSnapshot{InputTokens: 100})

	eng.Clear()
	assert.Empty(t, eng.Messages())
	assert.Equal(t, 0, eng.TotalUsage().InputTokens)
}

func TestEngine_SetMethods(t *testing.T) {
	eng := New(Options{})
	eng.SetModel("new-model")
	assert.Equal(t, "new-model", eng.opts.Model)

	eng.SetSystemPrompt("new prompt")
	assert.Equal(t, "new prompt", eng.opts.System)

	eng.SetMaxTurns(42)
	assert.Equal(t, 42, eng.opts.MaxTurns)
}
