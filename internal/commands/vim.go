package commands

import (
	"context"
	"fmt"
)

// vimCmd toggles vim mode.
type vimCmd struct{}

var _ Command = vimCmd{}

func (vimCmd) Name() string      { return "vim" }
func (vimCmd) ShortHelp() string { return "Toggle vim mode" }

func (vimCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	deps.Config.VimMode = !deps.Config.VimMode
	state := "off"
	if deps.Config.VimMode {
		state = "on"
	}
	return Result{Output: fmt.Sprintf("vim mode: %s", state)}, nil
}

// VimCmd returns a new vim command.
func VimCmd() Command { return vimCmd{} }
