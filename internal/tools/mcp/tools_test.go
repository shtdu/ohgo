package mcp

import (
	"context"
	"encoding/json"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	mcpmanager "github.com/shtdu/ohgo/internal/mcp"
	"github.com/shtdu/ohgo/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verify all four tools satisfy the tools.Tool interface at compile time.
var (
	_ tools.Tool = CallTool{}
	_ tools.Tool = ListResources{}
	_ tools.Tool = ReadResource{}
	_ tools.Tool = Auth{}
)

func TestCallTool_NilManager(t *testing.T) {
	tool := CallTool{Mgr: nil}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"server_name":"s","tool_name":"t","arguments":{}}`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "mcp manager not configured")
}

func TestCallTool_InvalidJSON(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := CallTool{Mgr: mgr}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestCallTool_Interface(t *testing.T) {
	var tool tools.Tool = CallTool{Mgr: nil}
	assert.Equal(t, "mcp_call_tool", tool.Name())
	assert.NotEmpty(t, tool.Description())
	assert.NotNil(t, tool.InputSchema())
}

func TestListResources_NilManager(t *testing.T) {
	tool := ListResources{Mgr: nil}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"server_name":"s"}`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "mcp manager not configured")
}

func TestListResources_InvalidJSON(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := ListResources{Mgr: mgr}
	result, err := tool.Execute(context.Background(), json.RawMessage(`not json`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestListResources_Interface(t *testing.T) {
	var tool tools.Tool = ListResources{Mgr: nil}
	assert.Equal(t, "mcp_list_resources", tool.Name())
	assert.NotEmpty(t, tool.Description())
	assert.NotNil(t, tool.InputSchema())
}

func TestReadResource_NilManager(t *testing.T) {
	tool := ReadResource{Mgr: nil}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"server_name":"s","uri":"file:///x"}`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "mcp manager not configured")
}

func TestReadResource_InvalidJSON(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := ReadResource{Mgr: mgr}
	result, err := tool.Execute(context.Background(), json.RawMessage(`broken`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestReadResource_Interface(t *testing.T) {
	var tool tools.Tool = ReadResource{Mgr: nil}
	assert.Equal(t, "mcp_read_resource", tool.Name())
	assert.NotEmpty(t, tool.Description())
	assert.NotNil(t, tool.InputSchema())
}

func TestAuth_NilManager(t *testing.T) {
	tool := Auth{Mgr: nil}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"server_name":"s","action":"status"}`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "mcp manager not configured")
}

func TestAuth_InvalidJSON(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := Auth{Mgr: mgr}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestAuth_UnsupportedAction(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := Auth{Mgr: mgr}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"server_name":"s","action":"login"}`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "unsupported action")
}

func TestAuth_StatusAllServers_NoServers(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := Auth{Mgr: mgr}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"server_name":"","action":"status"}`))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "No MCP servers connected")
}

func TestAuth_Interface(t *testing.T) {
	var tool tools.Tool = Auth{Mgr: nil}
	assert.Equal(t, "mcp_auth", tool.Name())
	assert.NotEmpty(t, tool.Description())
	assert.NotNil(t, tool.InputSchema())
}

// --- Execute paths with a non-nil (empty) MCP Manager ---

func TestCallTool_NotConnected(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := CallTool{Mgr: mgr}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"server_name":"s","tool_name":"t","arguments":{}}`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not connected")
}

func TestListResources_NotConnected(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := ListResources{Mgr: mgr}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"server_name":"s"}`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not connected")
}

func TestReadResource_NotConnected(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := ReadResource{Mgr: mgr}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"server_name":"s","uri":"file:///x"}`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not connected")
}

func TestExtractTextContent_TextOnly(t *testing.T) {
	contents := []mcpsdk.Content{
		&mcpsdk.TextContent{Text: "hello"},
		&mcpsdk.TextContent{Text: "world"},
	}
	got := extractTextContent(contents)
	assert.Equal(t, "hello\nworld", got)
}

func TestExtractTextContent_Empty(t *testing.T) {
	got := extractTextContent([]mcpsdk.Content{})
	assert.Equal(t, "[]", got)
}

func TestExtractTextContent_NilTextContent(t *testing.T) {
	// Slice with a nil Content entry should not panic and should fall back to JSON.
	var nilContent mcpsdk.Content = nil
	contents := []mcpsdk.Content{nilContent}
	got := extractTextContent(contents)
	// nil entry does not match *mcpsdk.TextContent, so text stays empty => JSON fallback
	assert.Equal(t, "[null]", got)
}

func TestAuth_StatusSpecific_NotConnected(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := Auth{Mgr: mgr}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{"server_name":"nonexistent","action":"status"}`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not connected")
}

// --- Context cancellation tests ---

func TestCallTool_ContextCancel(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := CallTool{Mgr: mgr}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := tool.Execute(ctx, json.RawMessage(`{"server_name":"s","tool_name":"t","arguments":{}}`))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestAuth_ContextCancel(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := Auth{Mgr: mgr}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := tool.Execute(ctx, json.RawMessage(`{"server_name":"s","action":"status"}`))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestListResources_ContextCancel(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := ListResources{Mgr: mgr}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := tool.Execute(ctx, json.RawMessage(`{"server_name":"s"}`))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestReadResource_ContextCancel(t *testing.T) {
	mgr := mcpmanager.NewManager()
	tool := ReadResource{Mgr: mgr}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := tool.Execute(ctx, json.RawMessage(`{"server_name":"s","uri":"file:///x"}`))
	assert.ErrorIs(t, err, context.Canceled)
}
