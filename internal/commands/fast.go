package commands

import (
	"context"
	"fmt"
)

// fastCmd toggles fast mode (maps to Verbose toggle).
type fastCmd struct{}

var _ Command = fastCmd{}

func (fastCmd) Name() string      { return "fast" }
func (fastCmd) ShortHelp() string { return "Toggle fast mode (verbose off/on)" }

func (fastCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	deps.Config.Verbose = !deps.Config.Verbose
	state := "off"
	if deps.Config.Verbose {
		state = "on"
	}
	return Result{Output: fmt.Sprintf("fast mode: verbose %s", state)}, nil
}

// FastState reports whether fast mode is active (verbose is on).
func FastState(deps *Deps) bool {
	return deps.Config.Verbose
}

// SetFastState sets fast mode.
func SetFastState(deps *Deps, on bool) {
	deps.Config.Verbose = on
}

// FastCmd returns a new fast command.
func FastCmd() Command { return fastCmd{} }
