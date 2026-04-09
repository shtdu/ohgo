package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// rewindCmd removes the last N turns from the conversation.
type rewindCmd struct{}

var _ Command = rewindCmd{}

func (rewindCmd) Name() string     { return "rewind" }
func (rewindCmd) ShortHelp() string { return "remove last N turns (default 1)" }

func (rewindCmd) Run(_ context.Context, args string, deps *Deps) (Result, error) {
	n := 1
	args = strings.TrimSpace(args)
	if args != "" {
		parsed, err := strconv.Atoi(args)
		if err != nil || parsed < 1 {
			return Result{}, fmt.Errorf("rewind: invalid argument %q, expected positive integer", args)
		}
		n = parsed
	}

	msgs := deps.Engine.Messages()
	if len(msgs) == 0 {
		return Result{Output: "rewind: no messages to rewind"}, nil
	}

	// Count pairs from the end: each "turn" is a user+assistant pair (and possibly tool_result messages).
	// We work backwards counting user messages as turn boundaries.
	removed := 0
	idx := len(msgs)
	for removed < n && idx > 0 {
		// Walk backwards to find the start of a turn (a user message)
		idx--
		for idx > 0 && msgs[idx].Role != "user" {
			idx--
		}
		removed++
	}

	if removed == 0 {
		return Result{Output: "rewind: no turns to remove"}, nil
	}

	trimmed := msgs[:idx]
	deps.Engine.LoadMessages(trimmed)

	return Result{
		Output: fmt.Sprintf("rewind: removed %d turn(s), %d messages remaining", removed, len(trimmed)),
	}, nil
}
