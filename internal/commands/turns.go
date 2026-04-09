package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// turnsCmd shows or sets the max turns limit.
type turnsCmd struct{}

var _ Command = turnsCmd{}

func (turnsCmd) Name() string     { return "turns" }
func (turnsCmd) ShortHelp() string { return "show or set max turns" }

func (turnsCmd) Run(_ context.Context, args string, deps *Deps) (Result, error) {
	args = strings.TrimSpace(args)
	if args == "" {
		// Show current state
		return Result{
			Output: fmt.Sprintf("max turns: %d (current: %d)", deps.Engine.MaxTurns(), deps.Engine.Turns()),
		}, nil
	}

	n, err := strconv.Atoi(args)
	if err != nil || n < 1 {
		return Result{}, fmt.Errorf("turns: invalid argument %q, expected positive integer", args)
	}

	deps.Engine.SetMaxTurns(n)
	return Result{
		Output: fmt.Sprintf("max turns set to %d", n),
	}, nil
}
