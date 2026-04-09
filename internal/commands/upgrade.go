package commands

import (
	"context"
	"fmt"
)

type upgradeCmd struct{}

var _ Command = upgradeCmd{}

func (upgradeCmd) Name() string { return "upgrade" }

func (upgradeCmd) ShortHelp() string {
	return "Show upgrade instructions"
}

func (upgradeCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	out := fmt.Sprintf("Upgrade:\n  go install github.com/shtdu/ohgo/cmd/og@latest\n\nCurrent version: %s", deps.Version)
	return Result{Output: out}, nil
}
