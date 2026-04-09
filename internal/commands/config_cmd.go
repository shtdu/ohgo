package commands

import (
	"context"
	"encoding/json"
	"fmt"
)

// configCmd shows the current configuration as JSON.
type configCmd struct{}

var _ Command = configCmd{}

func (configCmd) Name() string     { return "config" }
func (configCmd) ShortHelp() string { return "show current configuration" }

func (configCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	if deps.Config == nil {
		return Result{Output: "config: no configuration loaded"}, nil
	}

	data, err := json.MarshalIndent(deps.Config, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("config: marshal: %w", err)
	}

	return Result{Output: string(data)}, nil
}
