package message

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

// mockEmitter captures messages sent via the Emitter function.
type mockEmitter struct {
	messages []Message
	err      error
}

func (m *mockEmitter) emit(msg Message) error {
	m.messages = append(m.messages, msg)
	return m.err
}

func TestSendTool_Name(t *testing.T) {
	tool := SendTool{}
	assert.Equal(t, "send_message", tool.Name())
}

func TestSendTool_Description(t *testing.T) {
	tool := SendTool{}
	assert.Equal(t, "Send a message or notification", tool.Description())
}

func TestSendTool_InputSchema(t *testing.T) {
	tool := SendTool{}
	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
}

func TestSendTool_WithMockEmitter(t *testing.T) {
	emitter := &mockEmitter{}
	tool := SendTool{Emitter: emitter.emit}
	args, _ := json.Marshal(map[string]string{
		"content":   "Hello, world!",
		"recipient": "user1",
		"level":     "warning",
	})

	result, err := tool.Execute(context.Background(), args)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Message sent to user1")
	assert.Contains(t, result.Content, "level=warning")
	require.Len(t, emitter.messages, 1)
	assert.Equal(t, "Hello, world!", emitter.messages[0].Content)
	assert.Equal(t, "user1", emitter.messages[0].Recipient)
	assert.Equal(t, "warning", emitter.messages[0].Level)
	assert.False(t, emitter.messages[0].Timestamp.IsZero())
}

func TestSendTool_DefaultLevel(t *testing.T) {
	emitter := &mockEmitter{}
	tool := SendTool{Emitter: emitter.emit}
	args, _ := json.Marshal(map[string]string{
		"content": "Hello",
	})

	result, err := tool.Execute(context.Background(), args)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "level=info")
	require.Len(t, emitter.messages, 1)
	assert.Equal(t, "info", emitter.messages[0].Level)
}

func TestSendTool_NilEmitter(t *testing.T) {
	tool := SendTool{}
	args, _ := json.Marshal(map[string]string{"content": "test"})

	_, err := tool.Execute(context.Background(), args)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "emitter not configured")
}

func TestSendTool_EmptyContent(t *testing.T) {
	emitter := &mockEmitter{}
	tool := SendTool{Emitter: emitter.emit}
	args, _ := json.Marshal(map[string]string{"content": ""})

	result, err := tool.Execute(context.Background(), args)

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "content is required")
	assert.Empty(t, emitter.messages)
}

func TestSendTool_EmitterError(t *testing.T) {
	emitter := &mockEmitter{err: fmt.Errorf("network error")}
	tool := SendTool{Emitter: emitter.emit}
	args, _ := json.Marshal(map[string]string{"content": "test"})

	result, err := tool.Execute(context.Background(), args)

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "failed to send message")
	assert.Contains(t, result.Content, "network error")
}

func TestSendTool_InvalidJSON(t *testing.T) {
	emitter := &mockEmitter{}
	tool := SendTool{Emitter: emitter.emit}

	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

// Interface compliance check.
var _ tools.Tool = SendTool{}
