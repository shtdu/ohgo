package commands

import (
	"context"
)

// contextCmd shows the current system prompt.
type contextCmd struct{}

var _ Command = contextCmd{}

func (contextCmd) Name() string     { return "context" }
func (contextCmd) ShortHelp() string { return "show the system prompt" }

func (contextCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	prompt := deps.Engine.SystemPrompt()
	if prompt == "" {
		return Result{Output: "context: no system prompt configured"}, nil
	}
	return Result{Output: prompt}, nil
}
