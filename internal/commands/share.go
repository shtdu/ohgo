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

	// Build a simpler text representation
	type sharedMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	simple := make([]sharedMsg, 0, len(msgs))
	for _, msg := range msgs {
		simple = append(simple, sharedMsg{
			Role:    msg.Role,
			Content: msg.Text(),
		})
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
