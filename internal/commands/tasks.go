package commands

import (
	"context"
	"fmt"
	"strings"
)

type tasksCmd struct{}

var _ Command = tasksCmd{}

func (tasksCmd) Name() string      { return "tasks" }
func (tasksCmd) ShortHelp() string { return "list background tasks" }

func (tasksCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	if deps.Tasks == nil {
		return Result{Output: "tasks: no task manager"}, nil
	}

	taskList := deps.Tasks.List("")
	if len(taskList) == 0 {
		return Result{Output: "tasks: no tasks"}, nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Background tasks (%d):\n", len(taskList))
	for _, t := range taskList {
		fmt.Fprintf(&b, "  [%s] %s (%s)", string(t.Status), t.ID, t.Command)
		if t.Description != "" {
			fmt.Fprintf(&b, " - %s", t.Description)
		}
		fmt.Fprintln(&b)
	}
	return Result{Output: b.String()}, nil
}
