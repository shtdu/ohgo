package commands

import "context"

// keybindCmd shows current keybindings.
type keybindCmd struct{}

var _ Command = keybindCmd{}

func (keybindCmd) Name() string      { return "keybindings" }
func (keybindCmd) ShortHelp() string { return "Show keybindings" }

func (keybindCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: "keybindings: default keybindings active"}, nil
}

// KeybindCmd returns a new keybindings command.
func KeybindCmd() Command { return keybindCmd{} }
