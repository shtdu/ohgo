package engine

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/api"
)

func TestHasToolUses(t *testing.T) {
	tests := []struct {
		name string
		msg  api.Message
		want bool
	}{
		{
			"text only",
			api.NewUserTextMessage("hello"),
			false,
		},
		{
			"has tool_use",
			api.NewAssistantMessage([]api.ContentBlock{
				{Type: "text", Text: "let me check"},
				{Type: "tool_use", ID: "t1", Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)},
			}),
			true,
		},
		{
			"empty content",
			api.Message{Role: "assistant", Content: nil},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HasToolUses(tt.msg))
		})
	}
}

func TestExtractToolCalls(t *testing.T) {
	msg := api.NewAssistantMessage([]api.ContentBlock{
		{Type: "text", Text: "running commands"},
		{Type: "tool_use", ID: "t1", Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)},
		{Type: "tool_use", ID: "t2", Name: "read", Input: json.RawMessage(`{"path":"go.mod"}`)},
	})

	calls := ExtractToolCalls(msg)
	require.Len(t, calls, 2)
	assert.Equal(t, "t1", calls[0].ID)
	assert.Equal(t, "bash", calls[0].Name)
	assert.Equal(t, `{"command":"ls"}`, string(calls[0].Input))
	assert.Equal(t, "t2", calls[1].ID)
	assert.Equal(t, "read", calls[1].Name)
}

func TestBuildToolResultMessage(t *testing.T) {
	results := []ToolCallResult{
		{ToolUseID: "t1", Content: "file1.go\nfile2.go", IsError: false},
		{ToolUseID: "t2", Content: "permission denied", IsError: true},
	}

	msg := BuildToolResultMessage(results)
	assert.Equal(t, "user", msg.Role)
	require.Len(t, msg.Content, 2)

	assert.Equal(t, "tool_result", msg.Content[0].Type)
	assert.Equal(t, "t1", msg.Content[0].ToolUseID)
	assert.Equal(t, "file1.go\nfile2.go", msg.Content[0].Content)
	assert.False(t, msg.Content[0].IsError)

	assert.Equal(t, "tool_result", msg.Content[1].Type)
	assert.Equal(t, "t2", msg.Content[1].ToolUseID)
	assert.True(t, msg.Content[1].IsError)
}

func TestBuildToolResultMessage_Empty(t *testing.T) {
	msg := BuildToolResultMessage(nil)
	assert.Equal(t, "user", msg.Role)
	assert.Empty(t, msg.Content)
}
