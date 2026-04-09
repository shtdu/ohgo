package commands

import (
	"context"
)

type agentsCmd struct{}

var _ Command = agentsCmd{}

func (agentsCmd) Name() string      { return "agents" }
func (agentsCmd) ShortHelp() string { return "show subagent status (not yet available)" }

func (agentsCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: "agents: not yet implemented (coordinator subsystem pending)"}, nil
}
