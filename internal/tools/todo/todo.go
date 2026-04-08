// Package todo implements the todo_write tool for managing task lists.
package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

type todoItem struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type todoInput struct {
	Todos []todoItem `json:"todos"`
	Path  string     `json:"path"`
}

// TodoWriteTool formats todos as a markdown checklist and writes them to a file.
type TodoWriteTool struct{}

func (TodoWriteTool) Name() string { return "todo_write" }

func (TodoWriteTool) Description() string {
	return "Writes a task list as a markdown checklist file. Supports pending, in-progress, and completed statuses."
}

func (TodoWriteTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"todos": map[string]any{
				"type":  "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"subject": map[string]any{
							"type":        "string",
							"description": "Title of the task",
						},
						"description": map[string]any{
							"type":        "string",
							"description": "Additional details about the task",
						},
						"status": map[string]any{
							"type":        "string",
							"enum":        []string{"pending", "in_progress", "completed"},
							"description": "Current status of the task",
						},
					},
					"required": []string{"subject", "status"},
				},
			},
			"path": map[string]any{
				"type":        "string",
				"description": "File path to write todos to",
				"default":     "TODO.md",
			},
		},
		"required":             []string{"todos"},
		"additionalProperties": false,
	}
}

func (TodoWriteTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input todoInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	// Validate status values
	for _, todo := range input.Todos {
		switch todo.Status {
		case "pending", "in_progress", "completed":
			// valid
		default:
			return tools.Result{
				Content: fmt.Sprintf("invalid status %q: must be pending, in_progress, or completed", todo.Status),
				IsError: true,
			}, nil
		}
	}

	// Check context
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	// Resolve path — default to TODO.md
	path := input.Path
	if path == "" {
		path = "TODO.md"
	}
	path = tools.ResolvePath(path)

	// Build markdown content
	content := formatTodos(input.Todos)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return tools.Result{Content: fmt.Sprintf("cannot write todos: %v", err), IsError: true}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Wrote %d todo(s) to %s", len(input.Todos), path)}, nil
}

func formatTodos(todos []todoItem) string {
	var buf strings.Builder
	buf.WriteString("## Tasks\n\n")

	for _, todo := range todos {
		checkbox := statusCheckbox(todo.Status)
		line := checkbox + " " + todo.Subject
		if todo.Description != "" {
			line += " — " + todo.Description
		}
		buf.WriteString(line + "\n")
	}

	return buf.String()
}

func statusCheckbox(status string) string {
	switch status {
	case "pending":
		return "- [ ]"
	case "in_progress":
		return "- [~]"
	case "completed":
		return "- [x]"
	default:
		return "- [ ]"
	}
}
