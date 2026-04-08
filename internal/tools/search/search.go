// Package search implements the tool_search tool for finding tools by name or description.
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

type searchInput struct {
	Query string `json:"query"`
}

// SearchTool searches registered tools by name or description.
type SearchTool struct {
	Registry *tools.Registry
}

func (SearchTool) Name() string { return "tool_search" }

func (SearchTool) Description() string {
	return "Search available tools by name or description using a case-insensitive query."
}

func (SearchTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Search query to match against tool names and descriptions",
			},
		},
		"required":             []string{"query"},
		"additionalProperties": false,
	}
}

func (s SearchTool) Execute(_ context.Context, args json.RawMessage) (tools.Result, error) {
	var input searchInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.Query == "" {
		return tools.Result{Content: "query is required", IsError: true}, nil
	}

	if s.Registry == nil {
		return tools.Result{Content: "tool registry not available", IsError: true}, nil
	}

	lowerQuery := strings.ToLower(input.Query)
	var results []string

	for _, t := range s.Registry.List() {
		if strings.Contains(strings.ToLower(t.Name()), lowerQuery) ||
			strings.Contains(strings.ToLower(t.Description()), lowerQuery) {
			results = append(results, fmt.Sprintf("%s: %s", t.Name(), t.Description()))
		}
	}

	if len(results) == 0 {
		return tools.Result{Content: "(no matching tools found)"}, nil
	}

	return tools.Result{Content: strings.Join(results, "\n")}, nil
}
