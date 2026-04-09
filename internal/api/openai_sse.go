package api

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

// parseOpenAISSEStream reads an OpenAI-format SSE stream and emits normalized events.
// OpenAI format: each line is `data: {json}` or `data: [DONE]`.
// No `event:` prefix line (unlike Anthropic).
func parseOpenAISSEStream(reader io.Reader, ch chan<- StreamEvent) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var textAccum string
	var toolCalls []toolCallAccumulator
	var usage UsageSnapshot

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and non-data lines.
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end sentinel.
		if data == "[DONE]" {
			// Emit the complete message.
			blocks := assembleContentBlocks(textAccum, toolCalls)
			msg := NewAssistantMessage(blocks)
			ch <- StreamEvent{Type: "message_complete", Data: msg}
			ch <- StreamEvent{Type: "usage", Data: usage}
			return
		}

		var chunk openaiSSEChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		// Process choices.
		for _, choice := range chunk.Choices {
			// Accumulate text content.
			if choice.Delta.Content != "" {
				textAccum += choice.Delta.Content
				ch <- StreamEvent{Type: "text_delta", Data: choice.Delta.Content}
			}

			// Accumulate tool call fragments.
			for _, tc := range choice.Delta.ToolCalls {
				idx := tc.Index
				// Grow slice if needed.
				for len(toolCalls) <= idx {
					toolCalls = append(toolCalls, toolCallAccumulator{})
				}
				if tc.ID != "" {
					toolCalls[idx].id = tc.ID
				}
				if tc.Function.Name != "" {
					toolCalls[idx].name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					toolCalls[idx].args += tc.Function.Arguments
				}
			}

			// Check for finish.
			if choice.FinishReason != nil {
				for i := range toolCalls {
					toolCalls[i].finished = true
				}
			}
		}

		// Capture usage if present.
		if chunk.Usage != nil {
			usage.InputTokens = chunk.Usage.PromptTokens
			usage.OutputTokens = chunk.Usage.CompletionTokens
		}
	}
}
