package commands

import (
	"context"
)

type bridgeCmd struct{}

var _ Command = bridgeCmd{}

func (bridgeCmd) Name() string      { return "bridge" }
func (bridgeCmd) ShortHelp() string { return "bridge subsystem status (not yet available)" }

func (bridgeCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: "bridge: bridge subsystem not yet implemented"}, nil
}
