package brief

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestBriefTool_Name(t *testing.T) {
	assert.Equal(t, "brief", BriefTool{}.Name())
}

func TestBriefTool_ShortTextUnmodified(t *testing.T) {
	tool := BriefTool{}
	text := "Hello, world"
	args, _ := json.Marshal(map[string]any{"text": text})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, text, result.Content)
}

func TestBriefTool_TruncationWithEllipsis(t *testing.T) {
	tool := BriefTool{}
	text := strings.Repeat("a", 300)
	args, _ := json.Marshal(map[string]any{"text": text, "max_chars": 100})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(t, strings.HasSuffix(result.Content, "..."))
	// The truncated portion (before ...) should be <= 100 chars
	truncated := strings.TrimSuffix(result.Content, "...")
	assert.LessOrEqual(t, len(truncated), 100)
}

func TestBriefTool_StripTrailingWhitespace(t *testing.T) {
	tool := BriefTool{}
	// 10 chars of "hello" + 90 spaces = 100 chars, text is 101+ chars total
	text := "hello" + strings.Repeat("  ", 50) + "world"
	args, _ := json.Marshal(map[string]any{"text": text, "max_chars": 10})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	// The text[:10] is "hello     " — trailing spaces should be stripped before "..."
	assert.Equal(t, "hello...", result.Content)
}

func TestBriefTool_CustomMaxChars(t *testing.T) {
	tool := BriefTool{}
	text := strings.Repeat("x", 50)
	args, _ := json.Marshal(map[string]any{"text": text, "max_chars": 30})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, strings.Repeat("x", 30)+"...", result.Content)
}

func TestBriefTool_EmptyString(t *testing.T) {
	tool := BriefTool{}
	args, _ := json.Marshal(map[string]any{"text": ""})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "", result.Content)
}

func TestBriefTool_ExactlyMaxChars(t *testing.T) {
	tool := BriefTool{}
	text := strings.Repeat("a", 200)
	args, _ := json.Marshal(map[string]any{"text": text, "max_chars": 200})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, text, result.Content)
	assert.False(t, strings.HasSuffix(result.Content, "..."))
}

func TestBriefTool_OneCharOver(t *testing.T) {
	tool := BriefTool{}
	text := strings.Repeat("a", 201)
	args, _ := json.Marshal(map[string]any{"text": text, "max_chars": 200})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(t, strings.HasSuffix(result.Content, "..."))
	// Before "..." should be exactly 200 chars of 'a'
	assert.Equal(t, strings.Repeat("a", 200)+"...", result.Content)
}

func TestBriefTool_MinBound(t *testing.T) {
	tool := BriefTool{}
	text := strings.Repeat("b", 100)
	// max_chars below minimum (20) should be clamped to 20
	args, _ := json.Marshal(map[string]any{"text": text, "max_chars": 5})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, strings.Repeat("b", 20)+"...", result.Content)
}

func TestBriefTool_MaxBound(t *testing.T) {
	tool := BriefTool{}
	text := strings.Repeat("c", 3000)
	// max_chars above maximum (2000) should be clamped to 2000
	args, _ := json.Marshal(map[string]any{"text": text, "max_chars": 5000})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, strings.Repeat("c", 2000)+"...", result.Content)
}

func TestBriefTool_InvalidJSON(t *testing.T) {
	tool := BriefTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestBriefTool_DefaultMaxChars(t *testing.T) {
	tool := BriefTool{}
	text := strings.Repeat("d", 300)
	args, _ := json.Marshal(map[string]any{"text": text})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, strings.Repeat("d", 200)+"...", result.Content)
}

var _ tools.Tool = BriefTool{}
