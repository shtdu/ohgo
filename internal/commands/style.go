package commands

import (
	"context"
	"fmt"
	"strings"
)

// styleCmd shows or sets the output style.
type styleCmd struct{}

var _ Command = styleCmd{}

func (styleCmd) Name() string      { return "output-style" }
func (styleCmd) ShortHelp() string { return "Show or set the output style" }

func (styleCmd) Run(_ context.Context, args string, deps *Deps) (Result, error) {
	arg := strings.TrimSpace(args)
	if arg == "" {
		return Result{Output: fmt.Sprintf("output-style: %s", deps.Config.OutputStyle)}, nil
	}
	deps.Config.OutputStyle = arg
	return Result{Output: fmt.Sprintf("output-style: set to %s", arg)}, nil
}

// StyleCmd returns a new style command.
func StyleCmd() Command { return styleCmd{} }
