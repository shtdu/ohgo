package commands

import "context"

// versionCmd displays the current version.
type versionCmd struct{}

var _ Command = versionCmd{}

func (versionCmd) Name() string      { return "version" }
func (versionCmd) ShortHelp() string { return "Show the current version" }

func (versionCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	v := deps.Version
	if v == "" {
		v = "dev"
	}
	return Result{Output: v}, nil
}
