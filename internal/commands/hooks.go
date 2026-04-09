package commands

import (
	"context"
)

type hooksCmd struct{}

var _ Command = hooksCmd{}

func (hooksCmd) Name() string      { return "hooks" }
func (hooksCmd) ShortHelp() string { return "show configured hooks" }

func (hooksCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: "hooks: no hooks configured"}, nil
}
