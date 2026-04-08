package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shtdu/ohgo/internal/tasks"
	"github.com/shtdu/ohgo/internal/tools"
)

type listInput struct {
	Status string `json:"status,omitempty"`
}

// ListTool lists task records, optionally filtered by status.
type ListTool struct {
	Mgr *tasks.Manager
}

func (ListTool) Name() string { return "task_list" }

func (ListTool) Description() string {
	return "Lists background tasks, optionally filtered by status. Returns a formatted table of tasks."
}

func (ListTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status": map[string]any{
				"type":        "string",
				"description": "Filter by task status (pending, running, completed, failed, killed)",
			},
		},
		"additionalProperties": false,
	}
}

func (t ListTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	var input listInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	records := t.Mgr.List(tasks.Status(input.Status))

	if len(records) == 0 {
		msg := "No tasks found."
		if input.Status != "" {
			msg = fmt.Sprintf("No tasks found with status %q.", input.Status)
		}
		return tools.Result{Content: msg}, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%-12s %-12s %-12s %-40s %s\n", "ID", "Type", "Status", "Description", "CreatedAt")
	sb.WriteString(strings.Repeat("-", 100))
	sb.WriteString("\n")

	for _, rec := range records {
		desc := rec.Description
		if len(desc) > 38 {
			desc = desc[:35] + "..."
		}
		fmt.Fprintf(&sb, "%-12s %-12s %-12s %-40s %s\n",
			rec.ID,
			rec.Type,
			rec.Status,
			desc,
			rec.CreatedAt.Format("2006-01-02 15:04:05"),
	)
	}

	return tools.Result{Content: sb.String()}, nil
}
