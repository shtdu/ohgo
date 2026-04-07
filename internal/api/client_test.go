package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserTextMessage(t *testing.T) {
	msg := NewUserTextMessage("hello")
	assert.Equal(t, "user", msg.Role)
	require.Len(t, msg.Content, 1)
	assert.Equal(t, "text", msg.Content[0].Type)
	assert.Equal(t, "hello", msg.Content[0].Text)
}

func TestNewAssistantMessage(t *testing.T) {
	blocks := []ContentBlock{
		{Type: "text", Text: "hi"},
		{Type: "tool_use", ID: "t1", Name: "bash"},
	}
	msg := NewAssistantMessage(blocks)
	assert.Equal(t, "assistant", msg.Role)
	require.Len(t, msg.Content, 2)
}

func TestMessage_Text(t *testing.T) {
	msg := Message{
		Role: "assistant",
		Content: []ContentBlock{
			{Type: "text", Text: "hello "},
			{Type: "tool_use", ID: "t1", Name: "bash"},
			{Type: "text", Text: "world"},
		},
	}
	assert.Equal(t, "hello world", msg.Text())
}

func TestMessage_ToolUses(t *testing.T) {
	msg := Message{
		Role: "assistant",
		Content: []ContentBlock{
			{Type: "text", Text: "hi"},
			{Type: "tool_use", ID: "t1", Name: "bash"},
			{Type: "tool_use", ID: "t2", Name: "read"},
		},
	}
	uses := msg.ToolUses()
	require.Len(t, uses, 2)
	assert.Equal(t, "bash", uses[0].Name)
	assert.Equal(t, "read", uses[1].Name)
}

func TestMessage_ToolUses_Empty(t *testing.T) {
	msg := NewUserTextMessage("hello")
	assert.Empty(t, msg.ToolUses())
}

func TestContentBlock_JSONRoundTrip_Text(t *testing.T) {
	original := ContentBlock{Type: "text", Text: "hello world"}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ContentBlock
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "text", decoded.Type)
	assert.Equal(t, "hello world", decoded.Text)
}

func TestContentBlock_JSONRoundTrip_ToolUse(t *testing.T) {
	original := ContentBlock{
		Type:  "tool_use",
		ID:    "toolu_123",
		Name:  "bash",
		Input: json.RawMessage(`{"command":"ls"}`),
	}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ContentBlock
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "tool_use", decoded.Type)
	assert.Equal(t, "toolu_123", decoded.ID)
	assert.Equal(t, "bash", decoded.Name)
	assert.Equal(t, `{"command":"ls"}`, string(decoded.Input))
}

func TestContentBlock_JSONRoundTrip_ToolResult(t *testing.T) {
	original := ContentBlock{
		Type:      "tool_result",
		ToolUseID: "toolu_123",
		Content:   "file contents here",
		IsError:   false,
	}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ContentBlock
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "tool_result", decoded.Type)
	assert.Equal(t, "toolu_123", decoded.ToolUseID)
	assert.Equal(t, "file contents here", decoded.Content)
	assert.False(t, decoded.IsError)
}

func TestMessage_JSONRoundTrip(t *testing.T) {
	original := Message{
		Role: "user",
		Content: []ContentBlock{
			{Type: "text", Text: "hello"},
		},
	}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Message
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "user", decoded.Role)
	require.Len(t, decoded.Content, 1)
	assert.Equal(t, "hello", decoded.Content[0].Text)
}
