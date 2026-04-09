package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// passesCmd sets the number of reasoning passes.
type passesCmd struct{}

var _ Command = passesCmd{}

// reasoningPasses stores the current reasoning pass count.
var reasoningPasses = 1

func (passesCmd) Name() string      { return "passes" }
func (passesCmd) ShortHelp() string { return "Set reasoning passes (integer, default 1)" }

func (passesCmd) Run(_ context.Context, args string, _ *Deps) (Result, error) {
	arg := strings.TrimSpace(args)
	if arg == "" {
		return Result{Output: fmt.Sprintf("passes: %d", reasoningPasses)}, nil
	}
	n, err := strconv.Atoi(arg)
	if err != nil {
		return Result{}, fmt.Errorf("passes: invalid integer %q", arg)
	}
	if n < 1 {
		return Result{}, fmt.Errorf("passes: must be at least 1")
	}
	reasoningPasses = n
	return Result{Output: fmt.Sprintf("passes: set to %d", n)}, nil
}

// PassesCount returns the current reasoning pass count.
func PassesCount() int { return reasoningPasses }

// SetPassesCount sets the reasoning pass count.
func SetPassesCount(n int) { reasoningPasses = n }

// PassesCmd returns a new passes command.
func PassesCmd() Command { return passesCmd{} }
