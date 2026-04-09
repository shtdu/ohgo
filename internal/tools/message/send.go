// Package message implements the send_message tool for sending messages and notifications.
package message

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shtdu/ohgo/internal/tools"
)

// Message represents a message or notification sent by the agent.
type Message struct {
	Content   string    `json:"content"`
	Recipient string    `json:"recipient"`
	Level     string    `json:"level"`
	Timestamp time.Time `json:"timestamp"`
}

// SendTool sends a message or notification via the configured Emitter function.
type SendTool struct {
	Emitter func(msg Message) error
}

type sendInput struct {
	Content   string `json:"content"`
	Recipient string `json:"recipient,omitempty"`
	Level     string `json:"level,omitempty"`
}

func (SendTool) Name() string { return "send_message" }

func (SendTool) Description() string {
	return "Send a message or notification"
}

func (SendTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"content": map[string]any{
				"type":        "string",
				"description": "The content of the message to send",
			},
			"recipient": map[string]any{
				"type":        "string",
				"description": "The recipient of the message (optional)",
			},
			"level": map[string]any{
				"type":        "string",
				"description": "The message level, e.g. info, warning, error",
				"default":     "info",
			},
		},
		"required":             []string{"content"},
		"additionalProperties": false,
	}
}

func (s SendTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	if s.Emitter == nil {
		return tools.Result{}, fmt.Errorf("send_message: emitter not configured")
	}

	var input sendInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.Content == "" {
		return tools.Result{Content: "content is required", IsError: true}, nil
	}

	level := input.Level
	if level == "" {
		level = "info"
	}

	msg := Message{
		Content:   input.Content,
		Recipient: input.Recipient,
		Level:     level,
		Timestamp: time.Now(),
	}

	if err := s.Emitter(msg); err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to send message: %v", err), IsError: true}, nil
	}

	result := fmt.Sprintf("Message sent (level=%s)", msg.Level)
	if msg.Recipient != "" {
		result = fmt.Sprintf("Message sent to %s (level=%s)", msg.Recipient, msg.Level)
	}
	return tools.Result{Content: result}, nil
}
