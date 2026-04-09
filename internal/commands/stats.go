package commands

import (
	"context"
	"fmt"
)

// statsCmd shows turn and message counts.
type statsCmd struct{}

var _ Command = statsCmd{}

func (statsCmd) Name() string      { return "stats" }
func (statsCmd) ShortHelp() string { return "Show turn and message counts" }

func (statsCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	turns := deps.Engine.Turns()
	messages := len(deps.Engine.Messages())
	return Result{Output: formatKV(
		"Turns:", fmt.Sprintf("%d", turns),
		"Messages:", fmt.Sprintf("%d", messages),
	)}, nil
}
