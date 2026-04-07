package engine

import (
	"github.com/shtdu/ohgo/internal/api"
)

// ToolCallResult holds the result of a single tool execution.
type ToolCallResult struct {
	ToolUseID string
	Content   string
	IsError   bool
}

// HasToolUses checks if a message contains any tool_use blocks.
func HasToolUses(msg api.Message) bool {
	return len(msg.ToolUses()) > 0
}

// ExtractToolCalls extracts all tool_use blocks as ToolCall structs.
func ExtractToolCalls(msg api.Message) []api.ToolCall {
	blocks := msg.ToolUses()
	calls := make([]api.ToolCall, 0, len(blocks))
	for _, block := range blocks {
		calls = append(calls, api.ToolCall{
			ID:    block.ID,
			Name:  block.Name,
			Input: block.Input,
		})
	}
	return calls
}

// BuildToolResultMessage creates a user message with tool_result blocks.
func BuildToolResultMessage(results []ToolCallResult) api.Message {
	blocks := make([]api.ContentBlock, 0, len(results))
	for _, r := range results {
		blocks = append(blocks, api.ContentBlock{
			Type:      "tool_result",
			ToolUseID: r.ToolUseID,
			Content:   r.Content,
			IsError:   r.IsError,
		})
	}
	return api.Message{
		Role:    "user",
		Content: blocks,
	}
}
