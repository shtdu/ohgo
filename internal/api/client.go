// Package api defines the Client interface for communicating with LLM providers.
package api

import (
	"context"
	"encoding/json"
)

// Message represents a single message in a conversation.
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
type Client interface {
	// Stream sends a request and returns a channel of events.
	Stream(ctx context.Context, opts StreamOptions) (<-chan StreamEvent, error)
}
