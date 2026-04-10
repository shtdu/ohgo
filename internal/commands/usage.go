package commands

import (
	"context"
	"fmt"
)

// usageCmd shows a usage snapshot in a compact format.
type usageCmd struct{}

var _ Command = usageCmd{}

func (usageCmd) Name() string      { return "usage" }
func (usageCmd) ShortHelp() string { return "Show usage snapshot for this session" }

func (usageCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	usage := deps.Engine.TotalUsage()
	return Result{Output: fmt.Sprintf("Usage: %d input, %d output, %d total tokens (cache: %d read, %d created)",
		usage.InputTokens, usage.OutputTokens, usage.TotalTokens(),
		usage.CacheReadInputTokens, usage.CacheCreationInputTokens)}, nil
}
