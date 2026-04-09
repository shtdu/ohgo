package commands

import (
	"context"
)

// copyCmd is a placeholder for clipboard copy.
type copyCmd struct{}

var _ Command = copyCmd{}

func (copyCmd) Name() string     { return "copy" }
func (copyCmd) ShortHelp() string { return "copy last response to clipboard" }

func (copyCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: "copy: clipboard support not yet implemented"}, nil
}
