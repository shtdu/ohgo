// Package brief implements the brief tool for truncating text.
package brief

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

const (
	minMaxChars = 20
	maxMaxChars = 2000
	defaultMax  = 200
)

type briefInput struct {
	Text     string `json:"text"`
	MaxChars int    `json:"max_chars"`
}

// BriefTool truncates text to a specified maximum character count.
type BriefTool struct{}

func (BriefTool) Name() string { return "brief" }

func (BriefTool) Description() string {
	return "Truncate text to a specified maximum character count, appending ellipsis if truncated."
}

func (BriefTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"text": map[string]any{
				"type":        "string",
				"description": "The text to potentially truncate",
			},
			"max_chars": map[string]any{
				"type":        "integer",
				"description": "Maximum number of characters before truncation",
				"default":     defaultMax,
				"minimum":     minMaxChars,
				"maximum":     maxMaxChars,
			},
		},
		"required":             []string{"text"},
		"additionalProperties": false,
	}
}

func (BriefTool) Execute(_ context.Context, args json.RawMessage) (tools.Result, error) {
	var input briefInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.Text == "" {
		return tools.Result{Content: ""}, nil
	}

	maxChars := input.MaxChars
	if maxChars <= 0 {
		maxChars = defaultMax
	}
	if maxChars < minMaxChars {
		maxChars = minMaxChars
	}
	if maxChars > maxMaxChars {
		maxChars = maxMaxChars
	}

	if len(input.Text) <= maxChars {
		return tools.Result{Content: input.Text}, nil
	}

	truncated := strings.TrimRight(input.Text[:maxChars], " \t\n\r") + "..."
	return tools.Result{Content: truncated}, nil
}
