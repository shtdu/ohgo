package commands

import (
	"context"
	"fmt"
)

// statusCmd shows the current model, turns, and token usage.
type statusCmd struct{}

var _ Command = statusCmd{}

func (statusCmd) Name() string      { return "status" }
func (statusCmd) ShortHelp() string { return "Show current model, turns, and token usage" }

func (statusCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	model := deps.Engine.Model()
	turns := deps.Engine.Turns()
	maxTurns := deps.Engine.MaxTurns()
	usage := deps.Engine.TotalUsage()

	return Result{Output: formatKV(
		"Model:", model,
		"Turns:", fmt.Sprintf("%d / %d", turns, maxTurns),
		"Input tokens:", fmt.Sprintf("%d", usage.InputTokens),
		"Output tokens:", fmt.Sprintf("%d", usage.OutputTokens),
		"Total tokens:", fmt.Sprintf("%d", usage.TotalTokens()),
	)}, nil
}
