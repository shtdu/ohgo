package commands

import (
	"context"
	"fmt"

	"github.com/shtdu/ohgo/internal/engine"
)

// compactCmd forces conversation compaction.
type compactCmd struct{}

var _ Command = compactCmd{}

func (compactCmd) Name() string        { return "compact" }
func (compactCmd) ShortHelp() string    { return "force conversation compaction" }

func (compactCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	msgs := deps.Engine.Messages()
	if len(msgs) == 0 {
		return Result{Output: "compact: no messages to compact"}, nil
	}

	// Run microcompact
	result := engine.Microcompact(msgs, 5)
	if result.TokensSaved == 0 {
		return Result{Output: "compact: nothing to compact (all tool results are recent)"}, nil
	}

	deps.Engine.LoadMessages(result.Messages)

	return Result{
		Output: fmt.Sprintf("compact: cleared %d bytes from old tool results (%d messages remaining)",
			result.TokensSaved, len(result.Messages)),
	}, nil
}
