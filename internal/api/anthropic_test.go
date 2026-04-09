package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSSEServer creates a test server that serves SSE events.
func mockSSEServer(events []string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		if r.Header.Get("x-api-key") == "" {
			w.WriteHeader(401)
			fmt.Fprint(w, "missing api key")
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t := testing.Testing()
			_ = t
			return
		}

		for _, event := range events {
			fmt.Fprint(w, event)
			flusher.Flush()
		}
	}))
}

func TestAnthropicClient_TextStreaming(t *testing.T) {
	sseEvents := []string{
		"event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"usage\":{\"input_tokens\":10}}}\n\n",
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Hello\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\" world\"}}\n\n",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\"}\n\n",
		"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":5}}\n\n",
		"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
	}

	server := mockSSEServer(sseEvents)
	defer server.Close()

	client := NewAnthropicClient(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL+"/v1/messages"),
	)

	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "claude-sonnet-4-6",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	var textDeltas []string
	var gotComplete bool
	var gotUsage bool

	for event := range ch {
		switch event.Type {
		case "text_delta":
			textDeltas = append(textDeltas, event.Data.(string))
		case "message_complete":
			gotComplete = true
			msg := event.Data.(Message)
			assert.Equal(t, "assistant", msg.Role)
		case "usage":
			gotUsage = true
			usage := event.Data.(UsageSnapshot)
			assert.Equal(t, 10, usage.InputTokens)
			assert.Equal(t, 5, usage.OutputTokens)
		}
	}

	assert.Equal(t, []string{"Hello", " world"}, textDeltas)
	assert.True(t, gotComplete, "should receive message_complete event")
	assert.True(t, gotUsage, "should receive usage event")
}

func TestAnthropicClient_TextAccumulatedInContentBlock(t *testing.T) {
	// Verify that text_delta fragments are accumulated into contentBlock.Text
	// on content_block_stop (the behavior added by the textParts accumulator).
	sseEvents := []string{
		"event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"usage\":{\"input_tokens\":5}}}\n\n",
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Go\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\" \"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"rewritten\"}}\n\n",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\"}\n\n",
		"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":3}}\n\n",
		"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
	}

	server := mockSSEServer(sseEvents)
	defer server.Close()

	client := NewAnthropicClient(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL + "/v1/messages"),
	)

	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "claude-sonnet-4-6",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	for event := range ch {
		if event.Type == "message_complete" {
			msg := event.Data.(Message)
			require.Len(t, msg.Content, 1)
			assert.Equal(t, "Go rewritten", msg.Content[0].Text,
				"text delta fragments should be accumulated into contentBlock.Text")
			assert.Equal(t, "text", msg.Content[0].Type)
		}
	}
}

func TestAnthropicClient_MultipleContentBlocks_ResetTextParts(t *testing.T) {
	// Verify textParts resets between content blocks (text + tool_use + text).
	sseEvents := []string{
		"event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"usage\":{\"input_tokens\":10}}}\n\n",

		// Block 0: text
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Let me check\"}}\n\n",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\"}\n\n",

		// Block 1: tool_use with input_json_delta
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"content_block\":{\"type\":\"tool_use\",\"id\":\"tool_1\",\"name\":\"bash\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{\\\"command\\\":\\\"ls\\\"}\"}}\n\n",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\"}\n\n",

		// Block 2: text (continues after tool result)
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Done.\"}}\n\n",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\"}\n\n",

		"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":10}}\n\n",
		"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
	}

	server := mockSSEServer(sseEvents)
	defer server.Close()

	client := NewAnthropicClient(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL + "/v1/messages"),
	)

	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "claude-sonnet-4-6",
		MaxTokens: 200,
		Messages:  []Message{NewUserTextMessage("list files")},
	})
	require.NoError(t, err)

	for event := range ch {
		if event.Type == "message_complete" {
			msg := event.Data.(Message)
			require.Len(t, msg.Content, 3)

			// Block 0: text
			assert.Equal(t, "text", msg.Content[0].Type)
			assert.Equal(t, "Let me check", msg.Content[0].Text)

			// Block 1: tool_use
			assert.Equal(t, "tool_use", msg.Content[1].Type)
			assert.Equal(t, "bash", msg.Content[1].Name)
			assert.Equal(t, "tool_1", msg.Content[1].ID)
			assert.JSONEq(t, `{"command":"ls"}`, string(msg.Content[1].Input))

			// Block 2: text — verifies textParts was reset between blocks
			assert.Equal(t, "text", msg.Content[2].Type)
			assert.Equal(t, "Done.", msg.Content[2].Text,
				"textParts must reset between content blocks, not leak from block 0")
		}
	}
}

func TestAnthropicClient_AuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		fmt.Fprint(w, "invalid api key")
	}))
	defer server.Close()

	client := NewAnthropicClient(
		WithAPIKey("bad-key"),
		WithBaseURL(server.URL+"/v1/messages"),
		WithMaxRetries(0),
	)

	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "test",
		MaxTokens: 10,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	for event := range ch {
		if event.Type == "error" {
			assert.Contains(t, event.Data.(string), "authentication failed")
			return
		}
	}
	t.Fatal("expected error event")
}

func TestAnthropicClient_RetryOn429(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls <= 1 {
			w.WriteHeader(429)
			fmt.Fprint(w, "rate limited")
			return
		}
		// Success on second call
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		fmt.Fprint(w, "event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"usage\":{\"input_tokens\":5}}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"content_block\":{\"type\":\"text\",\"text\":\"ok\"}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event: content_block_stop\ndata: {\"type\":\"content_block_stop\"}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":2}}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	client := NewAnthropicClient(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL+"/v1/messages"),
		WithMaxRetries(2),
	)

	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "test",
		MaxTokens: 10,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	gotComplete := false
	for event := range ch {
		if event.Type == "message_complete" {
			gotComplete = true
		}
	}
	assert.True(t, gotComplete, "should succeed after retry")
	assert.Equal(t, 2, calls, "should have retried once")
}

func TestAnthropicClient_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Slow response
		time.Sleep(10 * time.Second)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	client := NewAnthropicClient(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL+"/v1/messages"),
		WithMaxRetries(0),
	)

	ch, err := client.Stream(ctx, StreamOptions{
		Model:     "test",
		MaxTokens: 10,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	for range ch {
		// drain
	}
	// Should complete without hanging
}

func TestAnthropicClient_BuildRequestBody(t *testing.T) {
	client := NewAnthropicClient(WithAPIKey("test"))
	body, err := client.buildRequestBody(StreamOptions{
		Model:     "claude-sonnet-4-6",
		MaxTokens: 100,
		System:    "you are helpful",
		Messages:  []Message{NewUserTextMessage("hello")},
		Tools: []ToolDef{
			{Name: "bash", Description: "run command", InputSchema: map[string]any{"type": "object"}},
		},
	})
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(body, &decoded))

	assert.Equal(t, "claude-sonnet-4-6", decoded["model"])
	assert.Equal(t, float64(100), decoded["max_tokens"])
	assert.Equal(t, true, decoded["stream"])
	assert.Equal(t, "you are helpful", decoded["system"])

	msgs, ok := decoded["messages"].([]any)
	require.True(t, ok)
	require.Len(t, msgs, 1)

	tools, ok := decoded["tools"].([]any)
	require.True(t, ok)
	require.Len(t, tools, 1)
}
