//go:build integration

package engine_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/api"
	"github.com/shtdu/ohgo/internal/engine"
	"github.com/shtdu/ohgo/internal/testutil"
)

// EARS: REQ-SM-001
func TestIntegration_Session_SaveAndRestore(t *testing.T) {
	f := testutil.NewFixture(t, testutil.FixtureConfig{
		Responses: []testutil.ResponseScript{
			{TextDeltas: []string{"hello"}, Usage: api.UsageSnapshot{InputTokens: 10, OutputTokens: 5}},
		},
	})

	_, stop := f.DrainEvents()
	err := f.Engine.Query(t.Context(), "hi")
	stop()
	require.NoError(t, err)

	// Save messages from first engine
	msgs := f.Engine.Messages()
	assert.NotEmpty(t, msgs)
	assert.Equal(t, "user", msgs[0].Role)

	// Create second engine and restore
	eng2 := engine.New(engine.Options{MaxTurns: 10})
	eng2.LoadMessages(msgs)

	// Verify restoration
	restored := eng2.Messages()
	require.Len(t, restored, len(msgs))
	assert.Equal(t, msgs[0].Content[0].Text, restored[0].Content[0].Text)
}

// EARS: REQ-SM-001
func TestIntegration_Session_MessagesAccessible(t *testing.T) {
	f := testutil.NewFixture(t, testutil.FixtureConfig{
		Responses: []testutil.ResponseScript{
			{TextDeltas: []string{"response"}, Usage: api.UsageSnapshot{InputTokens: 5, OutputTokens: 5}},
		},
	})

	_, stop := f.DrainEvents()
	err := f.Engine.Query(t.Context(), "test prompt")
	stop()
	require.NoError(t, err)

	msgs := f.Engine.Messages()
	// Should have user + assistant messages
	assert.GreaterOrEqual(t, len(msgs), 2)

	// First message is user prompt
	assert.Equal(t, "user", msgs[0].Role)
	assert.Equal(t, "test prompt", msgs[0].Content[0].Text)
}

// EARS: REQ-SM-001
func TestIntegration_Session_Clear(t *testing.T) {
	f := testutil.NewFixture(t, testutil.FixtureConfig{
		Responses: []testutil.ResponseScript{
			{TextDeltas: []string{"hi"}, Usage: api.UsageSnapshot{InputTokens: 5, OutputTokens: 5}},
		},
	})

	_, stop := f.DrainEvents()
	err := f.Engine.Query(t.Context(), "hi")
	stop()
	require.NoError(t, err)
	assert.NotEmpty(t, f.Engine.Messages())

	// Clear resets
	f.Engine.Clear()
	assert.Empty(t, f.Engine.Messages())
	assert.Equal(t, 0, f.Engine.Turns())
}

// EARS: REQ-SM-008
func TestIntegration_Compaction_ShouldCompact(t *testing.T) {
	// Under threshold: should not compact
	shortMsgs := []api.Message{api.NewUserTextMessage("short")}
	assert.False(t, engine.ShouldCompact(shortMsgs, 100000))

	// Over threshold: should compact
	longText := make([]byte, 400000)
	for i := range longText {
		longText[i] = 'a'
	}
	longMsgs := []api.Message{api.NewUserTextMessage(string(longText))}
	assert.True(t, engine.ShouldCompact(longMsgs, 100000))
}

// EARS: REQ-SM-008
func TestIntegration_Compaction_Microcompact(t *testing.T) {
	// Build a conversation with many tool results
	var messages []api.Message
	for i := range 10 {
		id := fmt.Sprintf("tool_%d", i)
		messages = append(messages,
			api.Message{
				Role: "assistant",
				Content: []api.ContentBlock{
					{Type: "tool_use", ID: id, Name: "bash", Input: json.RawMessage(`"echo hi"`)},
				},
			},
			api.Message{
				Role: "user",
				Content: []api.ContentBlock{
					{Type: "tool_result", ToolUseID: id, Content: strings.Repeat("output line ", 50)},
				},
			},
		)
	}

	result := engine.Microcompact(messages, 5)
	require.Len(t, result.Messages, len(messages))
	assert.Greater(t, result.TokensSaved, 0, "microcompact should save tokens")
}
