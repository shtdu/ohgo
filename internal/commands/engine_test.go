package commands

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/api"
	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/engine"
)

// testDeps creates a Deps with a real engine and config for testing.
func testDeps(t *testing.T) *Deps {
	t.Helper()
	eng := engine.New(engine.Options{
		Model:    "test-model",
		MaxTurns: 10,
	})
	cfg := config.DefaultSettings()
	return &Deps{
		Engine: eng,
		Config: &cfg,
	}
}

// loadTestMessages populates the engine with sample messages.
func loadTestMessages(eng *engine.Engine, count int) {
	msgs := make([]api.Message, count)
	for i := 0; i < count; i++ {
		if i%2 == 0 {
			msgs[i] = api.NewUserTextMessage("user message")
		} else {
			msgs[i] = api.NewAssistantMessage([]api.ContentBlock{
				{Type: "text", Text: "assistant message"},
			})
		}
	}
	eng.LoadMessages(msgs)
}

func TestCompactCmd_NoMessages(t *testing.T) {
	deps := testDeps(t)
	cmd := compactCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "no messages")
}

func testCompactCmd_WithData(t *testing.T) {
	deps := testDeps(t)
	// Create messages with tool_use and tool_result pairs that can be compacted
	msgs := []api.Message{
		api.NewUserTextMessage("check files"),
		api.NewAssistantMessage([]api.ContentBlock{
			{Type: "tool_use", ID: "t1", Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)},
		}),
		{Role: "user", Content: []api.ContentBlock{
			{Type: "tool_result", ToolUseID: "t1", Content: "file1.txt\nfile2.txt\nfile3.txt"},
		}},
		api.NewAssistantMessage([]api.ContentBlock{
			{Type: "tool_use", ID: "t2", Name: "bash", Input: json.RawMessage(`{"command":"ls -la"}`)},
		}),
		{Role: "user", Content: []api.ContentBlock{
			{Type: "tool_result", ToolUseID: "t2", Content: "detailed listing..."},
		}},
		api.NewUserTextMessage("summarize"),
		api.NewAssistantMessage([]api.ContentBlock{
			{Type: "text", Text: "here is the summary"},
		}),
	}
	deps.Engine.LoadMessages(msgs)

	cmd := compactCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "cleared")
}

func TestRewindCmd_DefaultOne(t *testing.T) {
	deps := testDeps(t)
	loadTestMessages(deps.Engine, 6) // 3 pairs of user+assistant

	cmd := rewindCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "removed 1 turn")
	assert.Equal(t, 4, len(deps.Engine.Messages()))
}

func TestRewindCmd_MultipleTurns(t *testing.T) {
	deps := testDeps(t)
	loadTestMessages(deps.Engine, 6) // 3 pairs

	cmd := rewindCmd{}
	res, err := cmd.Run(context.Background(), "2", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "removed 2 turn")
	assert.Equal(t, 2, len(deps.Engine.Messages()))
}

func TestRewindCmd_InvalidArg(t *testing.T) {
	deps := testDeps(t)
	cmd := rewindCmd{}
	_, err := cmd.Run(context.Background(), "abc", deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
}

func TestRewindCmd_NoMessages(t *testing.T) {
	deps := testDeps(t)
	cmd := rewindCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "no messages")
}

func TestTurnsCmd_Show(t *testing.T) {
	deps := testDeps(t)
	cmd := turnsCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "max turns: 10")
}

func TestTurnsCmd_Set(t *testing.T) {
	deps := testDeps(t)
	cmd := turnsCmd{}
	res, err := cmd.Run(context.Background(), "42", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "max turns set to 42")
	assert.Equal(t, 42, deps.Engine.MaxTurns())
}

func TestTurnsCmd_InvalidArg(t *testing.T) {
	deps := testDeps(t)
	cmd := turnsCmd{}
	_, err := cmd.Run(context.Background(), "abc", deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
}

func TestConfigCmd(t *testing.T) {
	deps := testDeps(t)
	cmd := configCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "claude-sonnet-4-6")
	assert.Contains(t, res.Output, "api_format")
}

func TestConfigCmd_Nil(t *testing.T) {
	deps := &Deps{}
	cmd := configCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "no configuration")
}

func TestExportCmd(t *testing.T) {
	deps := testDeps(t)
	loadTestMessages(deps.Engine, 4)

	cmd := exportCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "saved to")

	// Extract path from output and verify file exists
	parts := strings.SplitN(res.Output, " ", -1)
	path := parts[len(parts)-1]
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, len(data) > 0)

	// Verify it's valid JSON
	var parsed []api.Message
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, 4, len(parsed))
}

func TestExportCmd_NoMessages(t *testing.T) {
	deps := testDeps(t)
	cmd := exportCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "no messages")
}

func TestContextCmd(t *testing.T) {
	deps := testDeps(t)
	deps.Engine.SetSystemPrompt("you are a helpful assistant")

	cmd := contextCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Equal(t, "you are a helpful assistant", res.Output)
}

func TestContextCmd_Empty(t *testing.T) {
	deps := testDeps(t)
	cmd := contextCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "no system prompt")
}

func TestSummaryCmd(t *testing.T) {
	deps := testDeps(t)
	loadTestMessages(deps.Engine, 4)

	cmd := summaryCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "Messages: 4 total")
	assert.Contains(t, res.Output, "user")
	assert.Contains(t, res.Output, "assistant")
}

func TestSummaryCmd_Empty(t *testing.T) {
	deps := testDeps(t)
	cmd := summaryCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "no messages")
}

func TestContinueCmd(t *testing.T) {
	cmd := continueCmd{}
	res, err := cmd.Run(context.Background(), "", nil)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "not yet implemented")
}

func TestCopyCmd(t *testing.T) {
	cmd := copyCmd{}
	res, err := cmd.Run(context.Background(), "", nil)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "not yet implemented")
}

func TestTagAndResumeCmd(t *testing.T) {
	deps := testDeps(t)
	loadTestMessages(deps.Engine, 4)

	// Tag the session
	tag := tagCmd{}
	res, err := tag.Run(context.Background(), "test-session", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "saved")

	// Verify the tag file exists
	tagPath := filepath.Join(os.TempDir(), "ohgo-sessions", "test-session.json")
	_, err = os.ReadFile(tagPath)
	require.NoError(t, err)

	// Clear the engine
	deps.Engine.Clear()
	assert.Equal(t, 0, len(deps.Engine.Messages()))

	// Resume the session
	rsm := resumeCmd{}
	res, err = rsm.Run(context.Background(), "test-session", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "restored")

	// Verify messages were restored
	assert.Equal(t, 4, len(deps.Engine.Messages()))

	// Cleanup
	os.Remove(tagPath)
}

func TestTagCmd_ListEmpty(t *testing.T) {
	cmd := tagCmd{}
	res, err := cmd.Run(context.Background(), "", testDeps(t))
	require.NoError(t, err)
	// Either "no saved tags" or "Saved tags:" with entries from other tests
	assert.True(t,
		strings.Contains(res.Output, "no saved tags") || strings.Contains(res.Output, "Saved tags:"),
		res.Output)
}

func TestResumeCmd_NotFound(t *testing.T) {
	deps := testDeps(t)
	cmd := resumeCmd{}
	_, err := cmd.Run(context.Background(), "nonexistent-tag", deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestResumeCmd_NoArgs(t *testing.T) {
	deps := testDeps(t)
	cmd := resumeCmd{}
	_, err := cmd.Run(context.Background(), "", deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tag name required")
}

func TestSessionCmd(t *testing.T) {
	deps := testDeps(t)
	loadTestMessages(deps.Engine, 4)

	cmd := sessionCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "Session info")
	assert.Contains(t, res.Output, "Messages:")
	assert.Contains(t, res.Output, "Model:")
}

func TestShareCmd(t *testing.T) {
	deps := testDeps(t)
	loadTestMessages(deps.Engine, 4)

	cmd := shareCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "saved to")
}

func TestShareCmd_NoMessages(t *testing.T) {
	deps := testDeps(t)
	cmd := shareCmd{}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "no messages")
}

func TestTruncateText(t *testing.T) {
	assert.Equal(t, "hello", truncateText("hello", 10))
	assert.Equal(t, "hel...", truncateText("hello world", 6))
}

func TestCommandInterface(t *testing.T) {
	// Verify all commands satisfy the interface
	var cmds []Command = []Command{
		compactCmd{},
		contextCmd{},
		summaryCmd{},
		rewindCmd{},
		continueCmd{},
		turnsCmd{},
		configCmd{},
		exportCmd{},
		shareCmd{},
		copyCmd{},
		tagCmd{},
		resumeCmd{},
		sessionCmd{},
	}

	for _, cmd := range cmds {
		assert.NotEmpty(t, cmd.Name())
		assert.NotEmpty(t, cmd.ShortHelp())
	}
}
