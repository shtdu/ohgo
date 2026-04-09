package commands

import (
	"context"
	"fmt"
)

// costCmd shows token usage and estimated cost.
type costCmd struct{}

var _ Command = costCmd{}

func (costCmd) Name() string      { return "cost" }
func (costCmd) ShortHelp() string { return "Show token usage and estimated cost" }

func (costCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	usage := deps.Engine.TotalUsage()
	return Result{Output: formatKV(
		"Input tokens:", fmt.Sprintf("%d", usage.InputTokens),
		"Output tokens:", fmt.Sprintf("%d", usage.OutputTokens),
		"Total tokens:", fmt.Sprintf("%d", usage.TotalTokens()),
	)}, nil
}
