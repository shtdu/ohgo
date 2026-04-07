package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngineEvent_TypeSwitch(t *testing.T) {
	events := []EngineEvent{
		{Type: EventTextDelta, Data: AssistantTextDelta{Text: "hello"}},
		{Type: EventToolStarted, Data: ToolExecutionStarted{ToolName: "bash", ToolInput: `{"command":"ls"}`}},
		{Type: EventToolCompleted, Data: ToolExecutionCompleted{ToolName: "bash", Output: "file.go", IsError: false}},
		{Type: EventTurnComplete, Data: AssistantTurnComplete{InputTokens: 100, OutputTokens: 50}},
		{Type: EventError, Data: ErrorEvent{Message: "something broke", Recoverable: true}},
		{Type: EventStatus, Data: StatusEvent{Message: "loading..."}},
	}

	for _, event := range events {
		switch data := event.Data.(type) {
		case AssistantTextDelta:
			assert.Equal(t, EventTextDelta, event.Type)
			assert.Equal(t, "hello", data.Text)
		case ToolExecutionStarted:
			assert.Equal(t, EventToolStarted, event.Type)
			assert.Equal(t, "bash", data.ToolName)
		case ToolExecutionCompleted:
			assert.Equal(t, EventToolCompleted, event.Type)
			assert.Equal(t, "file.go", data.Output)
			assert.False(t, data.IsError)
		case AssistantTurnComplete:
			assert.Equal(t, EventTurnComplete, event.Type)
			assert.Equal(t, 100, data.InputTokens)
			assert.Equal(t, 50, data.OutputTokens)
		case ErrorEvent:
			assert.Equal(t, EventError, event.Type)
			assert.Equal(t, "something broke", data.Message)
			assert.True(t, data.Recoverable)
		case StatusEvent:
			assert.Equal(t, EventStatus, event.Type)
			assert.Equal(t, "loading...", data.Message)
		default:
			t.Fatalf("unexpected event data type: %T", data)
		}
	}
}

func TestAssistantTurnComplete_TokenCounts(t *testing.T) {
	event := AssistantTurnComplete{InputTokens: 500, OutputTokens: 200}
	assert.Equal(t, 500, event.InputTokens)
	assert.Equal(t, 200, event.OutputTokens)
}

func TestToolExecutionCompleted_ErrorFlag(t *testing.T) {
	event := ToolExecutionCompleted{ToolName: "bash", Output: "exit code 1", IsError: true}
	assert.True(t, event.IsError)
}
