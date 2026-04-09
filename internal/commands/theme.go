package commands

import (
	"context"
	"fmt"
	"strings"
)

// themeCmd shows or sets the UI theme.
type themeCmd struct{}

var _ Command = themeCmd{}

func (themeCmd) Name() string      { return "theme" }
func (themeCmd) ShortHelp() string { return "Show or set the UI theme" }

func (themeCmd) Run(_ context.Context, args string, deps *Deps) (Result, error) {
	arg := strings.TrimSpace(args)
	if arg == "" {
		return Result{Output: fmt.Sprintf("theme: %s", deps.Config.Theme)}, nil
	}
	deps.Config.Theme = arg
	return Result{Output: fmt.Sprintf("theme: set to %s", arg)}, nil
}

// ThemeCmd returns a new theme command.
func ThemeCmd() Command { return themeCmd{} }
