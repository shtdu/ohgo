package commands

import (
	"context"
	"fmt"
)

type reloadCmd struct{}

var _ Command = reloadCmd{}

func (reloadCmd) Name() string      { return "reload-plugins" }
func (reloadCmd) ShortHelp() string { return "re-discover plugins" }

func (reloadCmd) Run(ctx context.Context, _ string, deps *Deps) (Result, error) {
	if deps.Plugins == nil {
		return Result{Output: "reload-plugins: no plugin manager"}, nil
	}

	if err := deps.Plugins.Discover(ctx); err != nil {
		return Result{}, fmt.Errorf("reload-plugins: %w", err)
	}
	return Result{Output: "reload-plugins: plugin discovery complete"}, nil
}
