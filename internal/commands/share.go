package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// shareCmd exports the conversation with simpler formatting.
type shareCmd struct{}

var _ Command = shareCmd{}

func (shareCmd) Name() string     { return "share" }
func (shareCmd) ShortHelp() string { return "share conversation as formatted text" }

func (shareCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	msgs := deps.Engine.Messages()
	if len(msgs) == 0 {
		return Result{Output: "share: no messages to share"}, nil
	}

	// Build a readable representation that includes tool calls/results.
	type sharedBlock struct {
		Type  string `json:"type"`
		Text  string `json:"text,omitempty"`
		Name  string `json:"name,omitempty"`
		Input any    `json:"input,omitempty"`
	}
	type sharedMsg struct {
		Role    string        `json:"role"`
		Content []sharedBlock `json:"content"`
	}

	simple := make([]sharedMsg, 0, len(msgs))
	for _, msg := range msgs {
		blocks := make([]sharedBlock, 0, len(msg.Content))
		for _, b := range msg.Content {
			switch b.Type {
			case "text":
				if b.Text != "" {
					blocks = append(blocks, sharedBlock{Type: "text", Text: b.Text})
				}
			case "tool_use":
				var input any
				_ = json.Unmarshal(b.Input, &input)
				blocks = append(blocks, sharedBlock{Type: "tool_use", Name: b.Name, Input: input})
			case "tool_result":
				text := b.Content
				if text == "" {
					text = "(empty result)"
				}
				blocks = append(blocks, sharedBlock{Type: "tool_result", Name: b.ToolUseID, Text: text})
			}
		}
		if len(blocks) > 0 {
			simple = append(simple, sharedMsg{Role: msg.Role, Content: blocks})
		}
	}

	data, err := json.MarshalIndent(simple, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("share: marshal: %w", err)
	}

	f, err := os.CreateTemp("", "ohgo-share-*.json")
	if err != nil {
		return Result{}, fmt.Errorf("share: create temp file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return Result{}, fmt.Errorf("share: write: %w", err)
	}

	return Result{
		Output: fmt.Sprintf("share: conversation saved to %s", f.Name()),
	}, nil
}
