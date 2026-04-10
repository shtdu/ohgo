package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// anthropicSSEServer creates a test server that serves Anthropic-format SSE events.
func anthropicSSEServer(events []string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		if r.Header.Get("x-api-key") == "" {
			w.WriteHeader(401)
			_, _ = fmt.Fprint(w, `{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key"}}`)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}

		for _, event := range events {
			_, _ = fmt.Fprint(w, event)
			flusher.Flush()
		}
	}))
}

// Full SSE response for a simple text message: "Hello world"
func textSSEEvents() []string {
	return []string{
		"event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_1\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[],\"model\":\"claude-sonnet-4-6\",\"stop_reason\":null,\"stop_sequence\":null,\"usage\":{\"input_tokens\":10,\"output_tokens\":0,\"cache_creation_input_tokens\":0,\"cache_read_input_tokens\":0}}}\n\n",
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Hello\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\" world\"}}\n\n",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n",
		"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\",\"stop_sequence\":null},\"usage\":{\"output_tokens\":5}}\n\n",
		"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
	}
}

func TestAnthropicClient_TextStreaming(t *testing.T) {
	server := anthropicSSEServer(textSSEEvents())
	defer server.Close()

	client := NewAnthropicClient(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
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
	events := []string{
		"event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_2\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[],\"model\":\"test\",\"stop_reason\":null,\"stop_sequence\":null,\"usage\":{\"input_tokens\":5,\"output_tokens\":0,\"cache_creation_input_tokens\":0,\"cache_read_input_tokens\":0}}}\n\n",
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Go\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\" \"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"rewritten\"}}\n\n",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n",
		"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\",\"stop_sequence\":null},\"usage\":{\"output_tokens\":3}}\n\n",
		"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
	}

	server := anthropicSSEServer(events)
	defer server.Close()

	client := NewAnthropicClient(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
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

func TestAnthropicClient_MultipleContentBlocks(t *testing.T) {
	events := []string{
		"event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_3\",\"type\":\"message\",\"role\":\"assistant\",\"content\":[],\"model\":\"test\",\"stop_reason\":null,\"stop_sequence\":null,\"usage\":{\"input_tokens\":10,\"output_tokens\":0,\"cache_creation_input_tokens\":0,\"cache_read_input_tokens\":0}}}\n\n",

		// Block 0: text
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Let me check\"}}\n\n",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n",

		// Block 1: tool_use with input_json_delta
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":1,\"content_block\":{\"type\":\"tool_use\",\"id\":\"tool_1\",\"name\":\"bash\",\"input\":{}}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":1,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{\\\"command\\\":\\\"ls\\\"}\"}}\n\n",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":1}\n\n",

		// Block 2: text (continues after tool result)
		"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":2,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
		"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":2,\"delta\":{\"type\":\"text_delta\",\"text\":\"Done.\"}}\n\n",
		"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":2}\n\n",

		"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\",\"stop_sequence\":null},\"usage\":{\"output_tokens\":10}}\n\n",
		"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
	}

	server := anthropicSSEServer(events)
	defer server.Close()

	client := NewAnthropicClient(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
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
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key"}}`)
	}))
	defer server.Close()

	client := NewAnthropicClient(
		WithAPIKey("bad-key"),
		WithBaseURL(server.URL),
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
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"type":"error","error":{"type":"rate_limit_error","message":"rate limited"}}`)
			return
		}
		// Success on second call
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		for _, ev := range textSSEEvents() {
			_, _ = fmt.Fprint(w, ev)
			flusher.Flush()
		}
	}))
	defer server.Close()

	client := NewAnthropicClient(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
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
		time.Sleep(10 * time.Second)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	client := NewAnthropicClient(
		WithAPIKey("test-key"),
		WithBaseURL(server.URL),
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
