package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPServerConfig_IsEnabled(t *testing.T) {
	// Default (nil) means enabled.
	c := MCPServerConfig{}
	assert.True(t, c.IsEnabled())

	// Explicitly enabled.
	enabled := true
	c.Enabled = &enabled
	assert.True(t, c.IsEnabled())

	// Explicitly disabled.
	disabled := false
	c.Enabled = &disabled
	assert.False(t, c.IsEnabled())
}

func TestMCPSettings_JSONRoundTrip(t *testing.T) {
	enabled := true
	s := Settings{
		MCP: MCPSettings{
			Servers: []MCPServerConfig{
				{
					Name:      "test-server",
					Transport: "stdio",
					Command:   "test-mcp-server",
					Args:      []string{"--flag"},
					Enabled:   &enabled,
				},
				{
					Name:      "remote-server",
					Transport: "streamable_http",
					URL:       "http://localhost:8080/mcp",
					Headers:   map[string]string{"Authorization": "Bearer token"},
				},
			},
		},
	}

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var parsed Settings
	require.NoError(t, json.Unmarshal(data, &parsed))

	assert.Len(t, parsed.MCP.Servers, 2)
	assert.Equal(t, "test-server", parsed.MCP.Servers[0].Name)
	assert.Equal(t, "stdio", parsed.MCP.Servers[0].Transport)
	assert.Equal(t, "remote-server", parsed.MCP.Servers[1].Name)
	assert.Equal(t, "http://localhost:8080/mcp", parsed.MCP.Servers[1].URL)
}

func TestSettings_MCPDefaults(t *testing.T) {
	s := DefaultSettings()
	assert.Empty(t, s.MCP.Servers)
}

func TestMCPServerConfig_SSETransport(t *testing.T) {
	raw := `{
		"name": "sse-server",
		"transport": "sse",
		"url": "http://localhost:9090/sse"
	}`

	var c MCPServerConfig
	require.NoError(t, json.Unmarshal([]byte(raw), &c))
	assert.Equal(t, "sse-server", c.Name)
	assert.Equal(t, "sse", c.Transport)
	assert.Equal(t, "http://localhost:9090/sse", c.URL)
}
