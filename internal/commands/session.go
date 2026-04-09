package commands

import (
	"context"
	"fmt"
)

// sessionCmd shows session information.
type sessionCmd struct{}

var _ Command = sessionCmd{}

func (sessionCmd) Name() string     { return "session" }
func (sessionCmd) ShortHelp() string { return "show session info" }

func (sessionCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	msgs := deps.Engine.Messages()
	usage := deps.Engine.TotalUsage()

	// Count by role
	var userCount, assistantCount int
	for _, m := range msgs {
		switch m.Role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		}
	}

	out := "Session info:\n"
	out += fmt.Sprintf("  Temp dir:    %s\n", sessionDir())
	out += fmt.Sprintf("  Messages:    %d (%d user, %d assistant)\n",
		len(msgs), userCount, assistantCount)
	out += fmt.Sprintf("  Turns:       %d / %d\n", deps.Engine.Turns(), deps.Engine.MaxTurns())
	out += fmt.Sprintf("  Model:       %s\n", deps.Engine.Model())
	out += fmt.Sprintf("  Tokens:      %d in, %d out\n", usage.InputTokens, usage.OutputTokens)

	return Result{Output: out}, nil
}
