package commands

import (
	"context"
)

type mcpCmd struct{}

var _ Command = mcpCmd{}

func (mcpCmd) Name() string      { return "mcp" }
func (mcpCmd) ShortHelp() string { return "MCP client status (not yet available)" }

func (mcpCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: "mcp: MCP client not yet implemented"}, nil
}
