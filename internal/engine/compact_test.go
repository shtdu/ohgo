package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/api"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		messages []api.Message
		wantMin  int // minimum expected tokens
		wantMax  int // maximum expected tokens
	}{
		{
			name:     "empty",
			messages: nil,
			wantMin:  0,
			wantMax:  0,
		},
		{
			name: "text message",
			messages: []api.Message{
				api.NewUserTextMessage("hello world"),
			},
			wantMin: 1,
			wantMax: 20,
		},
		{
			name: "tool result",
			messages: []api.Message{
				{
					Role: "user",
					Content: []api.ContentBlock{
						{Type: "tool_result", Content: "some output text that is longer"},
					},
				},
			},
			wantMin: 5,
			wantMax: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.messages)
			assert.GreaterOrEqual(t, got, tt.wantMin)
			assert.LessOrEqual(t, got, tt.wantMax)
		})
	}
}

func TestMicrocompact_NoChangeWhenUnderThreshold(t *testing.T) {
	messages := []api.Message{
		{
			Role: "assistant",
			Content: []api.ContentBlock{
				{Type: "tool_use", ID: "1", Name: "bash", Input: json.RawMessage(`"ls"`)},
			},
		},
	}

	result := Microcompact(messages, 5)
	assert.Equal(t, messages, result.Messages)
	assert.Equal(t, 0, result.TokensSaved)
}

func TestMicrocompact_ClearsOldResults(t *testing.T) {
	// Create 7 tool_use + tool_result pairs (keep recent 5)
	messages := buildToolConversation(7)

	result := Microcompact(messages, 5)
	require.Len(t, result.Messages, len(messages))

	// First 2 tool results should be cleared
	clearedCount := 0
	for _, msg := range result.Messages {
		for _, block := range msg.Content {
			if block.Type == "tool_result" && block.Content == clearedMessage {
				clearedCount++
			}
		}
	}
	assert.Equal(t, 2, clearedCount)
	assert.Greater(t, result.TokensSaved, 0)
}

func TestMicrocompact_PreservesRecentResults(t *testing.T) {
	messages := buildToolConversation(7)
	result := Microcompact(messages, 5)

	// Last 5 tool results should NOT be cleared
	for _, msg := range result.Messages {
		for _, block := range msg.Content {
			if block.Type == "tool_result" {
				// Check that recent results still have original content
				if block.ToolUseID >= "tool_2" {
					assert.NotEqual(t, clearedMessage, block.Content,
						"recent tool result should not be cleared: %s", block.ToolUseID)
				}
			}
		}
	}
}

func TestMicrocompact_NonCompactableToolsNotCleared(t *testing.T) {
	messages := []api.Message{
		{
			Role: "assistant",
			Content: []api.ContentBlock{
				{Type: "tool_use", ID: "t1", Name: "ask_user", Input: json.RawMessage(`"question"`)},
			},
		},
		{
			Role: "user",
			Content: []api.ContentBlock{
				{Type: "tool_result", ToolUseID: "t1", Content: "user answer"},
			},
		},
	}

	result := Microcompact(messages, 0)
	assert.Equal(t, "user answer", result.Messages[1].Content[0].Content,
		"non-compactable tool result should not be cleared")
}

func TestShouldCompact(t *testing.T) {
	t.Run("under threshold", func(t *testing.T) {
		msgs := []api.Message{api.NewUserTextMessage("short")}
		assert.False(t, ShouldCompact(msgs, 100000))
	})

	t.Run("over threshold", func(t *testing.T) {
		// Create a large message
		longText := make([]byte, 400000)
		for i := range longText {
			longText[i] = 'a'
		}
		msgs := []api.Message{api.NewUserTextMessage(string(longText))}
		assert.True(t, ShouldCompact(msgs, 100000))
	})

	t.Run("zero context window", func(t *testing.T) {
		msgs := []api.Message{api.NewUserTextMessage("text")}
		assert.False(t, ShouldCompact(msgs, 0))
	})
}

func TestFullCompact_FewerThanPreserve(t *testing.T) {
	messages := []api.Message{
		api.NewUserTextMessage("msg1"),
		api.NewUserTextMessage("msg2"),
	}

	result, err := FullCompact(context.Background(), messages, nil, "", "", 10)
	require.NoError(t, err)
	assert.Equal(t, messages, result, "should return unchanged when under preserve count")
}

func TestFullCompact_WithMockAPI(t *testing.T) {
	// Mock SSE server that returns a summary
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: message_start\ndata: {\"message\":{\"usage\":{\"input_tokens\":10,\"output_tokens\":0}}}\n\n")
		fmt.Fprintf(w, "event: content_block_start\ndata: {\"type\":\"text\",\"text\":\"\"}\n\n")
		fmt.Fprintf(w, "event: content_block_delta\ndata: {\"delta\":{\"type\":\"text_delta\",\"text\":\"Summary of conversation\"}}\n\n")
		fmt.Fprintf(w, "event: content_block_stop\ndata: {}\n\n")
		fmt.Fprintf(w, "event: message_delta\ndata: {\"usage\":{\"output_tokens\":10}}\n\n")
		fmt.Fprintf(w, "event: message_stop\ndata: {}\n\n")
		w.(http.Flusher).Flush()
	}))
	defer server.Close()

	client := api.NewAnthropicClient(
		api.WithAPIKey("test-key"),
		api.WithBaseURL(server.URL),
		api.WithMaxRetries(0),
	)

	messages := buildToolConversation(10)
	result, err := FullCompact(context.Background(), messages, client, "test-model", "system", 6)
	require.NoError(t, err)

	// Should have summary + 6 preserved messages
	assert.Less(t, len(result), len(messages), "result should be shorter than input")
	assert.Equal(t, "user", result[0].Role)
	assert.Contains(t, result[0].Text(), "Conversation summary")
}

func TestFullCompact_APIClientError(t *testing.T) {
	// Server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := api.NewAnthropicClient(
		api.WithAPIKey("test-key"),
		api.WithBaseURL(server.URL),
		api.WithMaxRetries(0),
	)

	messages := buildToolConversation(10)
	result, err := FullCompact(context.Background(), messages, client, "test-model", "", 6)

	// Should return microcompacted messages on error (not the originals)
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, len(messages), len(result), "should preserve message count on error")
}

// buildToolConversation creates a conversation with n tool_use/tool_result pairs.
func buildToolConversation(n int) []api.Message {
	var messages []api.Message
	for i := range n {
		id := fmt.Sprintf("tool_%d", i)
		messages = append(messages,
			api.Message{
				Role: "assistant",
				Content: []api.ContentBlock{
					{Type: "tool_use", ID: id, Name: "bash", Input: json.RawMessage(`"echo hi"`)},
				},
			},
			api.Message{
				Role: "user",
				Content: []api.ContentBlock{
					{Type: "tool_result", ToolUseID: id, Content: fmt.Sprintf("output %d - some result text here", i)},
				},
			},
		)
	}
	return messages
}
