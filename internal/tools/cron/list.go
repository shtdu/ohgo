package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

// ListTool lists all cron jobs via the Manager.
type ListTool struct {
	Mgr *Manager
}

func (ListTool) Name() string { return "cron_list" }

func (ListTool) Description() string {
	return "Lists all configured cron jobs with their schedule, command, and enabled status."
}

func (ListTool) InputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":          map[string]any{},
		"additionalProperties": false,
	}
}

func (t ListTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	// Accept empty object or null; reject non-object input.
	if len(args) > 0 {
		var raw map[string]any
		if err := json.Unmarshal(args, &raw); err != nil {
			return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
		}
	}

	jobs := t.Mgr.List()
	if len(jobs) == 0 {
		return tools.Result{Content: "No cron jobs configured."}, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%-20s %-20s %-30s %s\n", "Name", "Schedule", "Command", "Enabled")
	sb.WriteString(strings.Repeat("-", 72))
	sb.WriteString("\n")
	for _, j := range jobs {
		fmt.Fprintf(&sb, "%-20s %-20s %-30s %t\n", j.Name, j.Schedule, j.Command, j.Enabled)
	}

	return tools.Result{Content: sb.String()}, nil
}
