package commands

import (
	"context"
)

// continueCmd is a placeholder for continuing tool execution.
type continueCmd struct{}

var _ Command = continueCmd{}

func (continueCmd) Name() string     { return "continue" }
func (continueCmd) ShortHelp() string { return "continue tool execution" }

func (continueCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: "continue: not yet implemented in standalone mode"}, nil
}
