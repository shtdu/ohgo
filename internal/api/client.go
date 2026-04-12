// Package api defines the Client interface for communicating with LLM providers.
package api

import (
	"context"
	"encoding/json"
)

// Message represents a single message in a conversation.
//
// Conversation history alternates between roles in a fixed pattern:
//
//	user:      "fix the bug"
//	assistant: [text, tool_use]          ← model responds, may request tools
//	user:      [tool_result]            ← engine provides tool output
//	assistant: "I've fixed it"          ← model continues
//	user:      "run the tests"
//	...
//
// ContentBlock types carry different fields depending on role and direction:
//   - text:        assistant → user display (Text field)
//   - tool_use:    assistant → engine (ID, Name, Input fields)
//   - tool_result: engine → API (ToolUseID, Content, IsError fields)
type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// NewUserTextMessage creates a user message with a single text block.
func NewUserTextMessage(text string) Message {
	return Message{
		Role: "user",
		Content: []ContentBlock{
			{Type: "text", Text: text},
		},
	}
}

// NewAssistantMessage creates an assistant message with the given content blocks.
func NewAssistantMessage(blocks []ContentBlock) Message {
	return Message{
		Role:    "assistant",
		Content: blocks,
	}
}

// Text returns the concatenated text from all text blocks.
func (m Message) Text() string {
	var b []byte
	for _, block := range m.Content {
		if block.Type == "text" {
			b = append(b, block.Text...)
		}
	}
	return string(b)
}

// ToolUses returns all tool_use content blocks.
func (m Message) ToolUses() []ContentBlock {
	var result []ContentBlock
	for _, block := range m.Content {
		if block.Type == "tool_use" {
			result = append(result, block)
		}
	}
	return result
}

// ContentBlock represents a typed content block in a message.
// Different fields are populated depending on Type:
//   - "text": Text is set
//   - "tool_use": ID, Name, Input are set
//   - "tool_result": ToolUseID, Content, IsError are set
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
}

// ToolCall represents a parsed tool_use request from the model.
type ToolCall struct {
	ID    string
	Name  string
	Input json.RawMessage
}

// StreamEvent represents a normalized event from the API client.
type StreamEvent struct {
	Type string
	Data any
}

// StreamOptions configures a streaming API request.
type StreamOptions struct {
	Model       string
	Messages    []Message
	Tools       []ToolDef
	MaxTokens   int
	Temperature float64
	System      string
}

// ToolDef describes a tool for the API request.
type ToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// Client is the interface for LLM API communication.
//
// Contract:
//   - Stream() returns immediately — events arrive on the returned channel asynchronously.
//   - The returned channel is always closed by the provider (EOF) or on error.
//   - Callers must drain the channel or cancel the context to avoid goroutine leaks.
//   - Implementations handle retry with exponential backoff internally. Callers see a
//     single logical stream — no retry logic is needed at the call site.
//   - API keys come from config/auth. Never pass credentials through StreamOptions.
//
// # Streaming Protocol
//
// Each provider uses a different SSE wire format, but all are normalized to the
// same StreamEvent channel. The engine doesn't know which provider it's using.
//
// Anthropic sends typed events (message_start, content_block_delta, message_stop).
// OpenAI and Copilot send newline-delimited JSON chunks terminated by [DONE].
// Both produce the same normalized events: text_delta, tool_use, message_complete,
// error, and usage.
type Client interface {
	// Stream sends a request and returns a channel of events.
	Stream(ctx context.Context, opts StreamOptions) (<-chan StreamEvent, error)
}
