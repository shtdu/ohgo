package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/shtdu/ohgo/internal/api"
)

// CompactableTools lists tools whose results can be cleared in microcompact.
var CompactableTools = map[string]bool{
	"read_file": true, "bash": true, "grep": true, "glob": true,
	"web_search": true, "web_fetch": true, "edit_file": true, "write_file": true,
}

const (
	clearedMessage       = "[Old tool result content cleared]"
	defaultKeepRecent    = 5
	autocompactBuffer    = 10000
	maxConsecutiveFails  = 3
)

// MicrocompactResult holds the result of a microcompact pass.
type MicrocompactResult struct {
	Messages    []api.Message
	TokensSaved int
}

// EstimateTokens estimates token count for a message slice.
// Uses a heuristic: ~4 chars per token, with 4/3 padding.
func EstimateTokens(messages []api.Message) int {
	totalChars := 0
	for _, msg := range messages {
		for _, block := range msg.Content {
			switch block.Type {
			case "text":
				totalChars += len(block.Text)
			case "tool_result":
				totalChars += len(block.Content)
			case "tool_use":
				totalChars += len(block.Name) + len(block.Input)
			}
		}
	}
	// Heuristic: ~3 chars per token (derived from ~4 chars/token with 4/3 padding for overhead)
	return totalChars / 3
}

// Microcompact clears old compactable tool results, keeping the most recent N.
func Microcompact(messages []api.Message, keepRecent int) MicrocompactResult {
	if keepRecent <= 0 {
		keepRecent = defaultKeepRecent
	}

	// Collect all tool_use IDs for compactable tools
	type toolUseInfo struct {
		id   string
		name string
	}
	var compactableIDs []toolUseInfo
	for _, msg := range messages {
		for _, block := range msg.Content {
			if block.Type == "tool_use" && CompactableTools[block.Name] {
				compactableIDs = append(compactableIDs, toolUseInfo{id: block.ID, name: block.Name})
			}
		}
	}

	if len(compactableIDs) <= keepRecent {
		return MicrocompactResult{Messages: messages, TokensSaved: 0}
	}

	// Mark older IDs for clearing (keep the last keepRecent)
	clearSet := make(map[string]bool, len(compactableIDs)-keepRecent)
	for i := 0; i < len(compactableIDs)-keepRecent; i++ {
		clearSet[compactableIDs[i].id] = true
	}

	// Walk messages and replace tool_result content for cleared IDs
	result := make([]api.Message, len(messages))
	tokensSaved := 0
	for i, msg := range messages {
		blocks := make([]api.ContentBlock, len(msg.Content))
		copy(blocks, msg.Content)
		for j, block := range blocks {
			if block.Type == "tool_result" && clearSet[block.ToolUseID] && block.Content != clearedMessage {
				tokensSaved += len(block.Content)
				blocks[j].Content = clearedMessage
			}
		}
		result[i] = api.Message{Role: msg.Role, Content: blocks}
	}

	return MicrocompactResult{Messages: result, TokensSaved: tokensSaved}
}

// ShouldCompact returns true if the estimated token count exceeds the threshold.
func ShouldCompact(messages []api.Message, contextWindow int) bool {
	if contextWindow <= 0 {
		return false
	}
	threshold := contextWindow - autocompactBuffer
	return EstimateTokens(messages) > threshold
}

// AutoCompactState tracks compaction across turns.
type AutoCompactState struct {
	Compacted           bool
	TurnCounter         int
	ConsecutiveFailures int
}

// FullCompact compacts messages by producing an LLM-generated summary.
// It microcompacts first, then splits into older (summarized) and newer (preserved).
func FullCompact(ctx context.Context, messages []api.Message, client api.Client, model string, systemPrompt string, preserveRecent int) ([]api.Message, error) {
	if preserveRecent <= 0 {
		preserveRecent = 6
	}

	// Microcompact first
	mcResult := Microcompact(messages, defaultKeepRecent)
	messages = mcResult.Messages

	if len(messages) <= preserveRecent {
		return messages, nil
	}

	// Split into older and newer
	splitIdx := len(messages) - preserveRecent
	older := messages[:splitIdx]
	newer := messages[splitIdx:]

	// Build conversation text for summarization
	var convText strings.Builder
	for _, msg := range older {
		fmt.Fprintf(&convText, "[%s]: ", msg.Role)
		for _, block := range msg.Content {
			switch block.Type {
			case "text":
				convText.WriteString(block.Text)
			case "tool_use":
				input := string(block.Input)
				if len(input) > 200 {
					input = input[:200] + "..."
				}
				fmt.Fprintf(&convText, "[tool:%s %s]", block.Name, input)
			case "tool_result":
				convText.WriteString(block.Content)
			}
		}
		convText.WriteString("\n")
	}

	// maxConsecutiveFails is used by the engine's auto-compact loop.
	_ = maxConsecutiveFails

	summaryPrompt := fmt.Sprintf(
		"Summarize the following conversation concisely, preserving key decisions, file paths, and outcomes. "+
			"This summary will replace the earlier conversation history.\n\n%s", convText.String())

	// Call API for summary (no tools)
	eventCh, err := client.Stream(ctx, api.StreamOptions{
		Model:     model,
		Messages:  []api.Message{api.NewUserTextMessage(summaryPrompt)},
		MaxTokens: 2048,
		System:    systemPrompt,
	})
	if err != nil {
		return messages, fmt.Errorf("compact summary call: %w", err)
	}

	var summary string
	for event := range eventCh {
		if data, ok := event.Data.(string); ok && event.Type == "text_delta" {
			summary += data
		}
	}

	if summary == "" {
		return messages, fmt.Errorf("compact: empty summary returned")
	}

	// Build result with proper alternating user/assistant message ordering.
	// The summary becomes a user message. If the first message in newer is also
	// a user message, merge the summary into it to avoid consecutive user messages.
	summaryText := fmt.Sprintf("[Conversation summary]\n%s\n\nThis session is being continued from a compacted conversation. "+
		"The summary above represents earlier conversation history.", summary)

	result := make([]api.Message, 0, 1+len(newer))

	if len(newer) > 0 && newer[0].Role == "user" {
		// Merge summary into the first user message to maintain alternating order
		merged := api.Message{
			Role: "user",
			Content: append([]api.ContentBlock{
				{Type: "text", Text: summaryText},
			}, newer[0].Content...),
		}
		result = append(result, merged)
		result = append(result, newer[1:]...)
	} else {
		result = append(result, api.NewUserTextMessage(summaryText))
		result = append(result, newer...)
	}
	return result, nil
}
