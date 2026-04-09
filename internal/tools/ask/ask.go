// Package ask implements the ask_user tool for prompting the user with questions.
package ask

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/tools"
)

// Prompter is the interface for asking the user a question interactively.
type Prompter interface {
	AskQuestion(ctx context.Context, question string, options []string, defaultVal string) (string, error)
}

// AskTool asks the user a question and returns their response.
type AskTool struct {
	Prompter Prompter
}

type askInput struct {
	Question string   `json:"question"`
	Options  []string `json:"options,omitempty"`
	Default  string   `json:"default,omitempty"`
}

func (AskTool) Name() string { return "ask_user" }

func (AskTool) Description() string {
	return "Ask the user a question and wait for their response"
}

func (AskTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"question": map[string]any{
				"type":        "string",
				"description": "The question to ask the user",
			},
			"options": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "List of options for the user to choose from (optional)",
			},
			"default": map[string]any{
				"type":        "string",
				"description": "Default value if the user does not provide one (optional)",
			},
		},
		"required":             []string{"question"},
		"additionalProperties": false,
	}
}

func (a AskTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	if a.Prompter == nil {
		return tools.Result{}, fmt.Errorf("ask_user: prompter not configured")
	}

	var input askInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.Question == "" {
		return tools.Result{Content: "question is required", IsError: true}, nil
	}

	answer, err := a.Prompter.AskQuestion(ctx, input.Question, input.Options, input.Default)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to get user response: %v", err), IsError: true}, nil
	}

	return tools.Result{Content: answer}, nil
}
