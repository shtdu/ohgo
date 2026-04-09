package commands

import (
	"context"
	"fmt"
	"strings"
)

// effortCmd sets reasoning effort level.
type effortCmd struct{}

var _ Command = effortCmd{}

// reasoningEffort stores the current reasoning effort level.
var reasoningEffort = "medium"

func (effortCmd) Name() string      { return "effort" }
func (effortCmd) ShortHelp() string { return "Set reasoning effort (low/medium/high)" }

func (effortCmd) Run(_ context.Context, args string, _ *Deps) (Result, error) {
	arg := strings.TrimSpace(strings.ToLower(args))
	if arg == "" {
		return Result{Output: fmt.Sprintf("effort: %s", reasoningEffort)}, nil
	}
	switch arg {
	case "low", "medium", "high":
		reasoningEffort = arg
		return Result{Output: fmt.Sprintf("effort: set to %s", arg)}, nil
	default:
		return Result{}, fmt.Errorf("effort: invalid value %q (use low, medium, or high)", arg)
	}
}

// EffortLevel returns the current reasoning effort level.
func EffortLevel() string { return reasoningEffort }

// SetEffortLevel sets the reasoning effort level.
func SetEffortLevel(level string) { reasoningEffort = level }

// EffortCmd returns a new effort command.
func EffortCmd() Command { return effortCmd{} }
