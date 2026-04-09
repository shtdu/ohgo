package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/config"
)

func TestConvertToOpenAIMessages_SystemPrompt(t *testing.T) {
	msgs := []Message{
		NewUserTextMessage("hello"),
	}
	result := convertToOpenAIMessages(msgs, "you are helpful")
	require.Len(t, result, 2)
	assert.Equal(t, "system", result[0].Role)
	assert.Equal(t, "you are helpful", result[0].Content)
	assert.Equal(t, "user", result[1].Role)
}

func TestConvertToOpenAIMessages_ToolResult(t *testing.T) {
	msgs := []Message{
		{
			Role: "user",
			Content: []ContentBlock{
				{Type: "tool_result", ToolUseID: "call_1", Content: "file contents"},
			},
		},
	}
	result := convertToOpenAIMessages(msgs, "")
	require.Len(t, result, 1)
	assert.Equal(t, "tool", result[0].Role)
	assert.Equal(t, "call_1", result[0].ToolCallID)
	assert.Equal(t, "file contents", result[0].Content)
}

func TestConvertToOpenAIMessages_AssistantWithToolUse(t *testing.T) {
	msgs := []Message{
		NewAssistantMessage([]ContentBlock{
			{Type: "text", Text: "let me check"},
			{Type: "tool_use", ID: "call_1", Name: "read_file", Input: json.RawMessage(`{"path":"/tmp/test"}`)},
		}),
	}
	result := convertToOpenAIMessages(msgs, "")
	require.Len(t, result, 1)
	assert.Equal(t, "assistant", result[0].Role)
	assert.Equal(t, "let me check", result[0].Content)
	require.Len(t, result[0].ToolCalls, 1)
	assert.Equal(t, "call_1", result[0].ToolCalls[0].ID)
	assert.Equal(t, "read_file", result[0].ToolCalls[0].Function.Name)
	assert.Equal(t, `{"path":"/tmp/test"}`, result[0].ToolCalls[0].Function.Arguments)
}

func TestConvertToOpenAITools(t *testing.T) {
	tools := []ToolDef{
		{
			Name:        "bash",
			Description: "Run a bash command",
			InputSchema: map[string]any{"type": "object"},
		},
	}
	result := convertToOpenAITools(tools)
	require.Len(t, result, 1)
	assert.Equal(t, "function", result[0].Type)
	assert.Equal(t, "bash", result[0].Function.Name)
}

func TestOpenAIClient_TextStreaming(t *testing.T) {
	sseData := `data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}

data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"content":" world"},"finish_reason":null}]}

data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	client := NewOpenAIClient(WithOpenAIAPIKey("test-key"), WithOpenAIBaseURL(server.URL))
	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	var events []StreamEvent
	for e := range ch {
		events = append(events, e)
	}

	// Should have: text_delta, text_delta, message_complete, usage
	var textParts []string
	var hasComplete, hasUsage bool
	for _, e := range events {
		switch e.Type {
		case "text_delta":
			textParts = append(textParts, e.Data.(string))
		case "message_complete":
			hasComplete = true
		case "usage":
			hasUsage = true
			u := e.Data.(UsageSnapshot)
			assert.Equal(t, 10, u.InputTokens)
			assert.Equal(t, 5, u.OutputTokens)
		}
	}

	assert.Equal(t, []string{"Hello", " world"}, textParts)
	assert.True(t, hasComplete)
	assert.True(t, hasUsage)
}

func TestOpenAIClient_ToolCallStreaming(t *testing.T) {
	sseData := `data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"role":"assistant","content":null,"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"bash","arguments":""}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"comma"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"nd\":\"ls\"}"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	client := NewOpenAIClient(WithOpenAIAPIKey("test-key"), WithOpenAIBaseURL(server.URL))
	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("run ls")},
	})
	require.NoError(t, err)

	var events []StreamEvent
	for e := range ch {
		events = append(events, e)
	}

	// Find message_complete event.
	var completeMsg Message
	for _, e := range events {
		if e.Type == "message_complete" {
			completeMsg = e.Data.(Message)
		}
	}

	toolUses := completeMsg.ToolUses()
	require.Len(t, toolUses, 1)
	assert.Equal(t, "call_1", toolUses[0].ID)
	assert.Equal(t, "bash", toolUses[0].Name)
	assert.Equal(t, `{"command":"ls"}`, string(toolUses[0].Input))
}

func TestOpenAIClient_AuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":{"message":"Invalid API key"}}`)
	}))
	defer server.Close()

	client := NewOpenAIClient(WithOpenAIAPIKey("bad-key"), WithOpenAIBaseURL(server.URL))
	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	var events []StreamEvent
	for e := range ch {
		events = append(events, e)
	}

	require.Len(t, events, 1)
	assert.Equal(t, "error", events[0].Type)
	assert.Contains(t, events[0].Data.(string), "authentication failed")
}

func TestOpenAIClient_RetryOn429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"error":{"message":"Rate limit exceeded"}}`)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, `data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"content":"OK"},"finish_reason":null}]}`+"\n\ndata: [DONE]\n")
	}))
	defer server.Close()

	client := NewOpenAIClient(
		WithOpenAIAPIKey("test-key"),
		WithOpenAIBaseURL(server.URL),
		WithOpenAIMaxRetries(3),
	)
	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	var events []StreamEvent
	for e := range ch {
		events = append(events, e)
	}

	assert.True(t, attempts >= 3, "should have retried")
	var hasText bool
	for _, e := range events {
		if e.Type == "text_delta" && e.Data.(string) == "OK" {
			hasText = true
		}
	}
	assert.True(t, hasText)
}

func TestOpenAIClient_ContextCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Slow response — never finishes.
		time.Sleep(10 * time.Second)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := NewOpenAIClient(WithOpenAIAPIKey("test-key"), WithOpenAIBaseURL(server.URL))
	ch, err := client.Stream(ctx, StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	// Channel should close with an error event.
	var events []StreamEvent
	for e := range ch {
		events = append(events, e)
	}
	require.Len(t, events, 1)
	assert.Equal(t, "error", events[0].Type)
}

func TestOpenAIClient_MixedTextAndToolCalls(t *testing.T) {
	sseData := `data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"role":"assistant","content":"Let me"},"finish_reason":null}]}

data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"content":" check."},"finish_reason":null}]}

data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"read_file","arguments":"{\"path\":\"/tmp\"}"}}]},"finish_reason":null}]}

data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	client := NewOpenAIClient(WithOpenAIAPIKey("test-key"), WithOpenAIBaseURL(server.URL))
	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("check files")},
	})
	require.NoError(t, err)

	var events []StreamEvent
	for e := range ch {
		events = append(events, e)
	}

	var completeMsg Message
	for _, e := range events {
		if e.Type == "message_complete" {
			completeMsg = e.Data.(Message)
		}
	}

	// Should have both text and tool_use blocks.
	assert.Equal(t, "Let me check.", completeMsg.Text())
	toolUses := completeMsg.ToolUses()
	require.Len(t, toolUses, 1)
	assert.Equal(t, "read_file", toolUses[0].Name)
}

func TestOpenAIClient_UserTextWithToolResult(t *testing.T) {
	// Test that a user message with both text and tool_result blocks converts correctly.
	msgs := []Message{
		NewAssistantMessage([]ContentBlock{
			{Type: "tool_use", ID: "call_1", Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)},
		}),
		{
			Role: "user",
			Content: []ContentBlock{
				{Type: "tool_result", ToolUseID: "call_1", Content: "file1.txt\nfile2.txt"},
			},
		},
	}
	result := convertToOpenAIMessages(msgs, "")
	// Should have: assistant with tool_call, tool with result
	require.Len(t, result, 2)
	assert.Equal(t, "assistant", result[0].Role)
	require.Len(t, result[0].ToolCalls, 1)
	assert.Equal(t, "tool", result[1].Role)
	assert.Equal(t, "call_1", result[1].ToolCallID)
}

func TestOpenAIClient_EmptySSEData(t *testing.T) {
	sseData := `data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	client := NewOpenAIClient(WithOpenAIAPIKey("test-key"), WithOpenAIBaseURL(server.URL))
	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	var events []StreamEvent
	for e := range ch {
		events = append(events, e)
	}

	// Should get message_complete even with empty content.
	var hasComplete bool
	for _, e := range events {
		if e.Type == "message_complete" {
			hasComplete = true
			msg := e.Data.(Message)
			assert.Equal(t, "", msg.Text())
		}
	}
	assert.True(t, hasComplete)
}

func TestOpenAIClient_SSELinesWithoutData(t *testing.T) {
	// Ensure non-data lines are skipped gracefully.
	sseData := `: comment line

data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"content":"hi"},"finish_reason":null}]}

event: ping
data: ignored

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	client := NewOpenAIClient(WithOpenAIAPIKey("test-key"), WithOpenAIBaseURL(server.URL))
	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	var events []StreamEvent
	for e := range ch {
		events = append(events, e)
	}

	// Only one text_delta from the valid data line.
	textCount := 0
	for _, e := range events {
		if e.Type == "text_delta" {
			textCount++
		}
	}
	assert.Equal(t, 1, textCount)
}

func TestOpenAIClient_MultipleToolCalls(t *testing.T) {
	sseData := `data: {"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"c1","type":"function","function":{"name":"bash","arguments":"{\"a\":"}},{"index":1,"id":"c2","type":"function","function":{"name":"read","arguments":"{\"b\":"}}]}}]}

data: {"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":"1}"}},{"index":1,"function":{"arguments":"2}"}}]}}]}

data: {"choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	client := NewOpenAIClient(WithOpenAIAPIKey("test-key"), WithOpenAIBaseURL(server.URL))
	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("run both")},
	})
	require.NoError(t, err)

	var completeMsg Message
	for e := range ch {
		if e.Type == "message_complete" {
			completeMsg = e.Data.(Message)
		}
	}

	toolUses := completeMsg.ToolUses()
	require.Len(t, toolUses, 2)
	assert.Equal(t, "bash", toolUses[0].Name)
	assert.Equal(t, `{"a":1}`, string(toolUses[0].Input))
	assert.Equal(t, "read", toolUses[1].Name)
	assert.Equal(t, `{"b":2}`, string(toolUses[1].Input))
}

func TestRegistry_OpenAIFactory(t *testing.T) {
	r := NewRegistry()
	cfg := &config.Settings{
		APIKey: "sk-openai-test",
		Profiles: map[string]config.ProviderProfile{
			"my-openai": {
				APIFormat: "openai",
				BaseURL:   "https://custom.openai.com/v1/chat/completions",
			},
		},
		ActiveProfile: "my-openai",
	}

	client, err := r.CreateClient(cfg, "")
	require.NoError(t, err)
	oc, ok := client.(*OpenAIClient)
	require.True(t, ok)
	assert.Equal(t, "sk-openai-test", oc.apiKey)
	assert.Equal(t, "https://custom.openai.com/v1/chat/completions", oc.baseURL)
}

// Ensure OpenAIClient satisfies Client interface.
func TestOpenAIClient_Interface(t *testing.T) {
	var _ Client = (*OpenAIClient)(nil)
}

// Ensure strings.Builder is used where applicable.
func init() {
	_ = strings.Builder{}
}
