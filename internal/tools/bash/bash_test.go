package bash

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestBashTool_NameAndSchema(t *testing.T) {
	tool := BashTool{}
	assert.Equal(t, "bash", tool.Name())
	assert.Contains(t, tool.Description(), "shell")

	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
	required := schema["required"].([]string)
	assert.Contains(t, required, "command")
}

func TestBashTool_EchoCommand(t *testing.T) {
	tool := BashTool{}
	args, _ := json.Marshal(map[string]string{"command": "echo hello world"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "hello world")
}

func TestBashTool_NonZeroExitCode(t *testing.T) {
	tool := BashTool{}
	args, _ := json.Marshal(map[string]string{"command": "exit 42"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "[exit code: 42]")
}

func TestBashTool_StderrCapture(t *testing.T) {
	tool := BashTool{}
	args, _ := json.Marshal(map[string]string{
		"command": "echo stderr-msg >&2 && echo stdout-msg",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "stderr-msg")
	assert.Contains(t, result.Content, "stdout-msg")
}

func TestBashTool_Timeout(t *testing.T) {
	tool := BashTool{}
	args, _ := json.Marshal(map[string]any{
		"command":         "sleep 60",
		"timeout_seconds": 1,
	})

	start := time.Now()
	result, err := tool.Execute(context.Background(), args)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "timed out")
	assert.Contains(t, result.Content, "1s")
	// The command should complete close to the 1s timeout, not 60s.
	assert.Less(t, elapsed, 5*time.Second)
}

func TestBashTool_ContextCancellation(t *testing.T) {
	tool := BashTool{}
	ctx, cancel := context.WithCancel(context.Background())

	args, _ := json.Marshal(map[string]string{"command": "sleep 60"})

	done := make(chan tools.Result, 1)
	go func() {
		r, _ := tool.Execute(ctx, args)
		done <- r
	}()

	// Cancel after a short delay.
	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case result := <-done:
		// Context cancellation returns an error (ctx.Err()), not a Result.
		// But it depends on timing — could also get the cancel result.
		_ = result
	case <-time.After(5 * time.Second):
		t.Fatal("context cancellation did not terminate the command")
	}
}

func TestBashTool_LargeOutputTruncation(t *testing.T) {
	tool := BashTool{}
	// Generate output larger than 12000 characters.
	args, _ := json.Marshal(map[string]string{
		"command": "python3 -c \"print('x' * 15000)\"",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.LessOrEqual(t, len(result.Content), 12000+len(truncatedSuffix))
	assert.Contains(t, result.Content, "...[truncated]...")
}

func TestBashTool_WorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	tool := BashTool{}
	args, _ := json.Marshal(map[string]string{
		"command": "pwd",
	})

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)
	os.Chdir(dir)

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	// The shell should inherit the working directory.
	assert.Contains(t, result.Content, filepath.Base(dir))
}

func TestBashTool_InvalidJSON(t *testing.T) {
	tool := BashTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestBashTool_EmptyCommand(t *testing.T) {
	tool := BashTool{}
	args, _ := json.Marshal(map[string]string{"command": ""})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "command is required")
}

func TestBashTool_MultiLineOutput(t *testing.T) {
	tool := BashTool{}
	args, _ := json.Marshal(map[string]string{
		"command": "echo line1 && echo line2 && echo line3",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "line1")
	assert.Contains(t, result.Content, "line2")
	assert.Contains(t, result.Content, "line3")
}

// Verify the tool satisfies the interface.
var _ tools.Tool = BashTool{}
