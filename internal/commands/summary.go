package commands

import (
	"context"
	"fmt"
)

// summaryCmd shows a conversation summary.
type summaryCmd struct{}

var _ Command = summaryCmd{}

func (summaryCmd) Name() string     { return "summary" }
func (summaryCmd) ShortHelp() string { return "show conversation summary" }

func (summaryCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	msgs := deps.Engine.Messages()
	if len(msgs) == 0 {
		return Result{Output: "summary: no messages in conversation"}, nil
	}

	// Count messages by role
	var userCount, assistantCount, toolResultCount int
	for _, msg := range msgs {
		switch msg.Role {
		case "user":
			userCount++
			for _, block := range msg.Content {
				if block.Type == "tool_result" {
					toolResultCount++
				}
			}
		case "assistant":
			assistantCount++
		}
	}

	out := fmt.Sprintf("Messages: %d total (%d user, %d assistant, %d tool results)\n",
		len(msgs), userCount, assistantCount, toolResultCount)

	// Show last few messages briefly
	out += "\nRecent messages:\n"
	start := len(msgs)
	if start > 5 {
		start = len(msgs) - 5
	}
	for i := start; i < len(msgs); i++ {
		msg := msgs[i]
		text := truncateText(msg.Text(), 80)
		out += fmt.Sprintf("  [%d] %s: %s\n", i+1, msg.Role, text)
	}

	return Result{Output: out}, nil
}

// truncateText truncates text to maxLen with ellipsis.
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
