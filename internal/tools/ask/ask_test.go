package ask

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

// mockPrompter captures AskQuestion calls.
type mockPrompter struct {
	answer string
	err    error
}

func (m *mockPrompter) AskQuestion(_ context.Context, _ string, _ []string, _ string) (string, error) {
	return m.answer, m.err
}

func TestAskTool_Name(t *testing.T) {
	tool := AskTool{}
	assert.Equal(t, "ask_user", tool.Name())
}

func TestAskTool_Description(t *testing.T) {
	tool := AskTool{}
	assert.Equal(t, "Ask the user a question and wait for their response", tool.Description())
}

func TestAskTool_InputSchema(t *testing.T) {
	tool := AskTool{}
	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
}

func TestAskTool_WithMockPrompter(t *testing.T) {
	prompter := &mockPrompter{answer: "yes"}
	tool := AskTool{Prompter: prompter}
	args, _ := json.Marshal(map[string]any{
		"question": "Do you want to continue?",
		"options":  []string{"yes", "no"},
		"default":  "yes",
	})

	result, err := tool.Execute(context.Background(), args)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "yes", result.Content)
}

func TestAskTool_NilPrompter(t *testing.T) {
	tool := AskTool{}
	args, _ := json.Marshal(map[string]string{"question": "test?"})

	_, err := tool.Execute(context.Background(), args)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "prompter not configured")
}

func TestAskTool_EmptyQuestion(t *testing.T) {
	prompter := &mockPrompter{}
	tool := AskTool{Prompter: prompter}
	args, _ := json.Marshal(map[string]string{"question": ""})

	result, err := tool.Execute(context.Background(), args)

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "question is required")
}

func TestAskTool_PrompterError(t *testing.T) {
	prompter := &mockPrompter{err: fmt.Errorf("cancelled")}
	tool := AskTool{Prompter: prompter}
	args, _ := json.Marshal(map[string]string{"question": "test?"})

	result, err := tool.Execute(context.Background(), args)

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "failed to get user response")
	assert.Contains(t, result.Content, "cancelled")
}

func TestAskTool_InvalidJSON(t *testing.T) {
	prompter := &mockPrompter{}
	tool := AskTool{Prompter: prompter}

	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

// Interface compliance check.
var _ tools.Tool = AskTool{}
