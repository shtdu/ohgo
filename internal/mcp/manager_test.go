package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/config"
)

func TestManager_New(t *testing.T) {
	m := NewManager()
	assert.Empty(t, m.List())
}

func TestManager_Get_NotFound(t *testing.T) {
	m := NewManager()
	_, ok := m.Get("nonexistent")
	assert.False(t, ok)
}

func TestManager_CloseAll_Empty(t *testing.T) {
	m := NewManager()
	assert.NoError(t, m.CloseAll())
}

func TestManager_CloseAll_Idempotent(t *testing.T) {
	m := NewManager()
	assert.NoError(t, m.CloseAll())
	assert.NoError(t, m.CloseAll())
}

func TestManager_Disconnect_NotFound(t *testing.T) {
	m := NewManager()
	err := m.Disconnect("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestManager_ConnectAll_SkipsDisabled(t *testing.T) {
	m := NewManager()
	disabled := false
	servers := []config.MCPServerConfig{
		{
			Name:      "disabled-server",
			Transport: "stdio",
			Command:   "nonexistent-command",
			Enabled:   &disabled,
		},
	}
	err := m.ConnectAll(context.Background(), servers)
	assert.NoError(t, err)
	assert.Empty(t, m.List())
}

func TestManager_CallTool_NotConnected(t *testing.T) {
	m := NewManager()
	_, err := m.CallTool(context.Background(), "missing", "tool", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestManager_ListTools_NotConnected(t *testing.T) {
	m := NewManager()
	_, err := m.ListTools(context.Background(), "missing")
	require.Error(t, err)
}

func TestManager_ListResources_NotConnected(t *testing.T) {
	m := NewManager()
	_, err := m.ListResources(context.Background(), "missing")
	require.Error(t, err)
}

func TestManager_ReadResource_NotConnected(t *testing.T) {
	m := NewManager()
	_, err := m.ReadResource(context.Background(), "missing", "resource://test")
	require.Error(t, err)
}

func TestNewTransport_Stdio(t *testing.T) {
	cfg := config.MCPServerConfig{
		Transport: "stdio",
		Command:   "echo",
		Args:      []string{"hello"},
	}
	_, err := newTransport(cfg)
	assert.NoError(t, err)
}

func TestNewTransport_Stdio_MissingCommand(t *testing.T) {
	cfg := config.MCPServerConfig{
		Transport: "stdio",
	}
	_, err := newTransport(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires a command")
}

func TestNewTransport_StreamableHTTP(t *testing.T) {
	cfg := config.MCPServerConfig{
		Transport: "streamable_http",
		URL:       "http://localhost:8080/mcp",
	}
	_, err := newTransport(cfg)
	assert.NoError(t, err)
}

func TestNewTransport_StreamableHTTP_MissingURL(t *testing.T) {
	cfg := config.MCPServerConfig{
		Transport: "streamable_http",
	}
	_, err := newTransport(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires a URL")
}

func TestNewTransport_SSE(t *testing.T) {
	cfg := config.MCPServerConfig{
		Transport: "sse",
		URL:       "http://localhost:9090/sse",
	}
	_, err := newTransport(cfg)
	assert.NoError(t, err)
}

func TestNewTransport_SSE_MissingURL(t *testing.T) {
	cfg := config.MCPServerConfig{
		Transport: "sse",
	}
	_, err := newTransport(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires a URL")
}

func TestNewTransport_Unknown(t *testing.T) {
	cfg := config.MCPServerConfig{
		Transport: "websocket",
	}
	_, err := newTransport(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown transport")
}
