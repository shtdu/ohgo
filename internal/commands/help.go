package commands

import (
	"context"
	"fmt"
	"strings"
)

// helpCmd lists all registered commands with their short help text.
type helpCmd struct{}

var _ Command = helpCmd{}

func (helpCmd) Name() string      { return "help" }
func (helpCmd) ShortHelp() string { return "List available commands" }

func (helpCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	if deps.CmdReg == nil {
		return Result{Output: "No commands registered."}, nil
	}

	cmds := deps.CmdReg.List()
	if len(cmds) == 0 {
		return Result{Output: "No commands registered."}, nil
	}

	// Find the longest command name for alignment
	maxName := 0
	for _, c := range cmds {
		if len(c.Name()) > maxName {
			maxName = len(c.Name())
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Available commands:\n\n")
	for _, c := range cmds {
		fmt.Fprintf(&b, "  /%-*s  %s\n", maxName, c.Name(), c.ShortHelp())
	}

	return Result{Output: b.String()}, nil
}
