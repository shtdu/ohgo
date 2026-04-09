package api

import "encoding/json"

// OpenAI-format message types for serialization.

type openaiMessage struct {
	Role       string            `json:"role"`
	Content    string            `json:"content,omitempty"`
	ToolCalls  []openaiToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
}

type openaiToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"` // always "function"
	Function openaiToolFunction `json:"function"`
}

type openaiToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openaiToolDef struct {
	Type     string              `json:"type"` // always "function"
	Function openaiToolFunctionDef `json:"function"`
}

type openaiToolFunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// convertToOpenAIMessages translates internal Messages to OpenAI chat completion format.
func convertToOpenAIMessages(messages []Message, system string) []openaiMessage {
	var result []openaiMessage

	// System message goes first.
	if system != "" {
		result = append(result, openaiMessage{Role: "system", Content: system})
	}

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			result = append(result, convertUserMessage(msg)...)
		case "assistant":
			result = append(result, convertAssistantMessage(msg))
		}
	}
	return result
}

// convertUserMessage handles user messages which may contain text and tool_result blocks.
func convertUserMessage(msg Message) []openaiMessage {
	var result []openaiMessage
	var textParts []string

	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			textParts = append(textParts, block.Text)
		case "tool_result":
			// OpenAI uses separate messages with role="tool" for tool results.
			result = append(result, openaiMessage{
				Role:       "tool",
				ToolCallID: block.ToolUseID,
				Content:    block.Content,
			})
		}
	}

	// If there are text parts, create a user message.
	if len(textParts) > 0 {
		combined := ""
		for i, t := range textParts {
			if i > 0 {
				combined += "\n"
			}
			combined += t
		}
		// Prepend user text message before tool results.
		result = append([]openaiMessage{{Role: "user", Content: combined}}, result...)
	}

	return result
}

// convertAssistantMessage handles assistant messages with text and tool_use blocks.
func convertAssistantMessage(msg Message) openaiMessage {
	om := openaiMessage{Role: "assistant"}

	var textParts []string
	var toolCalls []openaiToolCall

	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			textParts = append(textParts, block.Text)
		case "tool_use":
			args := "{}"
			if len(block.Input) > 0 {
				args = string(block.Input)
			}
			toolCalls = append(toolCalls, openaiToolCall{
				ID:   block.ID,
				Type: "function",
				Function: openaiToolFunction{
					Name:      block.Name,
					Arguments: args,
				},
			})
		}
	}

	for _, t := range textParts {
		om.Content += t
	}
	if len(toolCalls) > 0 {
		om.ToolCalls = toolCalls
	}

	return om
}

// convertToOpenAITools translates ToolDefs to OpenAI function-calling format.
func convertToOpenAITools(tools []ToolDef) []openaiToolDef {
	result := make([]openaiToolDef, len(tools))
	for i, t := range tools {
		result[i] = openaiToolDef{
			Type: "function",
			Function: openaiToolFunctionDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		}
	}
	return result
}

// openaiRequest is the full request body for the OpenAI chat completions API.
type openaiRequest struct {
	Model       string           `json:"model"`
	Messages    []openaiMessage  `json:"messages"`
	Tools       []openaiToolDef  `json:"tools,omitempty"`
	MaxTokens   int              `json:"max_tokens"`
	Temperature float64          `json:"temperature,omitempty"`
	Stream      bool             `json:"stream"`
}

// openaiSSEChunk represents a single SSE data payload from OpenAI.
type openaiSSEChunk struct {
	ID      string           `json:"id"`
	Choices []openaiChoice   `json:"choices"`
	Usage   *openaiUsageInfo `json:"usage,omitempty"`
}

type openaiChoice struct {
	Index        int             `json:"index"`
	Delta        openaiDelta     `json:"delta"`
	FinishReason *string         `json:"finish_reason"`
}

type openaiDelta struct {
	Role      string            `json:"role,omitempty"`
	Content   string            `json:"content,omitempty"`
	ToolCalls []openaiToolDelta `json:"tool_calls,omitempty"`
}

type openaiToolDelta struct {
	Index    int                 `json:"index"`
	ID       string              `json:"id,omitempty"`
	Type     string              `json:"type,omitempty"`
	Function openaiToolFuncDelta `json:"function"`
}

type openaiToolFuncDelta struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type openaiUsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// toolCallAccumulator tracks partial tool call arguments across SSE chunks.
type toolCallAccumulator struct {
	id       string
	name     string
	args     string
	finished bool
}

// assembleContentBlocks builds ContentBlocks from accumulated text and tool calls.
func assembleContentBlocks(text string, toolCalls []toolCallAccumulator) []ContentBlock {
	var blocks []ContentBlock
	if text != "" {
		blocks = append(blocks, ContentBlock{Type: "text", Text: text})
	}
	for _, tc := range toolCalls {
		blocks = append(blocks, ContentBlock{
			Type:  "tool_use",
			ID:    tc.id,
			Name:  tc.name,
			Input: json.RawMessage(tc.args),
		})
	}
	return blocks
}
