package commands

import (
	"context"
	"fmt"
	"strings"
)

// planCmd toggles or shows plan mode.
type planCmd struct{}

var _ Command = planCmd{}

func (planCmd) Name() string      { return "plan" }
func (planCmd) ShortHelp() string { return "Toggle plan mode" }

func (planCmd) Run(_ context.Context, args string, deps *Deps) (Result, error) {
	arg := strings.TrimSpace(strings.ToLower(args))

	if arg == "" {
		// Toggle current state
		if deps.Config.OutputStyle == "plan" {
			deps.Config.OutputStyle = "default"
			return Result{Output: "Plan mode disabled."}, nil
		}
		deps.Config.OutputStyle = "plan"
		return Result{Output: "Plan mode enabled."}, nil
	}

	switch arg {
	case "on", "true", "yes", "enable":
		deps.Config.OutputStyle = "plan"
		return Result{Output: "Plan mode enabled."}, nil
	case "off", "false", "no", "disable":
		deps.Config.OutputStyle = "default"
		return Result{Output: "Plan mode disabled."}, nil
	default:
		return Result{}, fmt.Errorf("unknown argument %q; use \"on\" or \"off\"", arg)
	}
}
