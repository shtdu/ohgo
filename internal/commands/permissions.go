package commands

import (
	"context"
	"fmt"
)

// permissionsCmd shows the current permission mode and settings.
type permissionsCmd struct{}

var _ Command = permissionsCmd{}

func (permissionsCmd) Name() string      { return "permissions" }
func (permissionsCmd) ShortHelp() string { return "Show current permission mode" }

func (permissionsCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	mode := deps.Config.Permission.Mode
	if mode == "" {
		mode = "default"
	}
	output := fmt.Sprintf("Permission mode: %s", mode)

	if len(deps.Config.Permission.AllowedTools) > 0 {
		output += fmt.Sprintf("\nAllowed tools: %v", deps.Config.Permission.AllowedTools)
	}
	if len(deps.Config.Permission.DeniedTools) > 0 {
		output += fmt.Sprintf("\nDenied tools: %v", deps.Config.Permission.DeniedTools)
	}

	return Result{Output: output}, nil
}
