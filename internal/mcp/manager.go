// Package mcp manages MCP client connections to external MCP servers
// using the official Model Context Protocol Go SDK.
package mcp

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shtdu/ohgo/internal/config"
)

// ServerConn wraps an active MCP client session.
type ServerConn struct {
	Name    string
	Config  config.MCPServerConfig
	Session *mcp.ClientSession
}

// Manager manages connections to multiple MCP servers.
type Manager struct {
	mu          sync.RWMutex
	connections map[string]*ServerConn
}

// NewManager creates an MCP connection manager.
func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*ServerConn),
	}
}

// ConnectAll connects to all enabled MCP servers from settings.
func (m *Manager) ConnectAll(ctx context.Context, servers []config.MCPServerConfig) error {
	var firstErr error
	for _, s := range servers {
		if !s.IsEnabled() {
			continue
		}
		if err := m.Connect(ctx, s); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("connect %q: %w", s.Name, err)
		}
	}
	return firstErr
}

// Connect establishes a connection to a single MCP server.
func (m *Manager) Connect(ctx context.Context, cfg config.MCPServerConfig) error {
	transport, err := newTransport(cfg)
	if err != nil {
		return err
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "ohgo", Version: "0.1.0"}, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return fmt.Errorf("mcp connect %q: %w", cfg.Name, err)
	}

	m.mu.Lock()
	// Close existing connection if replacing.
	if existing, ok := m.connections[cfg.Name]; ok {
		existing.Session.Close()
	}
	m.connections[cfg.Name] = &ServerConn{
		Name:    cfg.Name,
		Config:  cfg,
		Session: session,
	}
	m.mu.Unlock()

	return nil
}

// Disconnect closes a connection by server name.
func (m *Manager) Disconnect(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, ok := m.connections[name]
	if !ok {
		return fmt.Errorf("mcp: server %q not connected", name)
	}
	delete(m.connections, name)
	return conn.Session.Close()
}

// CloseAll shuts down all connections.
func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for name, conn := range m.connections {
		if err := conn.Session.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(m.connections, name)
	}
	return firstErr
}

// Get returns a connected server session by name.
func (m *Manager) Get(name string) (*ServerConn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.connections[name]
	return conn, ok
}

// List returns all active connections.
func (m *Manager) List() []*ServerConn {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ServerConn, 0, len(m.connections))
	for _, conn := range m.connections {
		result = append(result, conn)
	}
	return result
}

// CallTool invokes a tool on a specific MCP server.
func (m *Manager) CallTool(ctx context.Context, serverName, toolName string, args map[string]any) (*mcp.CallToolResult, error) {
	conn, ok := m.Get(serverName)
	if !ok {
		return nil, fmt.Errorf("mcp: server %q not connected", serverName)
	}
	return conn.Session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
}

// ListTools returns all tools from a specific MCP server.
func (m *Manager) ListTools(ctx context.Context, serverName string) (*mcp.ListToolsResult, error) {
	conn, ok := m.Get(serverName)
	if !ok {
		return nil, fmt.Errorf("mcp: server %q not connected", serverName)
	}
	return conn.Session.ListTools(ctx, nil)
}

// ListResources returns all resources from a specific MCP server.
func (m *Manager) ListResources(ctx context.Context, serverName string) (*mcp.ListResourcesResult, error) {
	conn, ok := m.Get(serverName)
	if !ok {
		return nil, fmt.Errorf("mcp: server %q not connected", serverName)
	}
	return conn.Session.ListResources(ctx, nil)
}

// ReadResource reads a resource from a specific MCP server.
func (m *Manager) ReadResource(ctx context.Context, serverName, uri string) (*mcp.ReadResourceResult, error) {
	conn, ok := m.Get(serverName)
	if !ok {
		return nil, fmt.Errorf("mcp: server %q not connected", serverName)
	}
	return conn.Session.ReadResource(ctx, &mcp.ReadResourceParams{URI: uri})
}

// newTransport creates an MCP transport from config.
func newTransport(cfg config.MCPServerConfig) (mcp.Transport, error) {
	switch cfg.Transport {
	case "stdio", "":
		cmd := cfg.Command
		if cmd == "" {
			return nil, fmt.Errorf("mcp: stdio transport requires a command")
		}
		execCmd := exec.Command(cmd, cfg.Args...)
		// Inject custom environment variables.
		if len(cfg.Env) > 0 {
			execCmd.Env = append(execCmd.Environ(), envMapToSlice(cfg.Env)...)
		}
		return &mcp.CommandTransport{Command: execCmd}, nil
	case "streamable_http":
		if cfg.URL == "" {
			return nil, fmt.Errorf("mcp: streamable_http transport requires a URL")
		}
		return &mcp.StreamableClientTransport{
			Endpoint:   cfg.URL,
			HTTPClient: httpClientWithHeaders(cfg.Headers),
		}, nil
	case "sse":
		if cfg.URL == "" {
			return nil, fmt.Errorf("mcp: sse transport requires a URL")
		}
		return &mcp.SSEClientTransport{
			Endpoint:   cfg.URL,
			HTTPClient: httpClientWithHeaders(cfg.Headers),
		}, nil
	default:
		return nil, fmt.Errorf("mcp: unknown transport %q", cfg.Transport)
	}
}

// envMapToSlice converts a map of environment variables to "KEY=VALUE" slices.
func envMapToSlice(env map[string]string) []string {
	s := make([]string, 0, len(env))
	for k, v := range env {
		s = append(s, k+"="+v)
	}
	return s
}

// httpClientWithHeaders creates an *http.Client that injects the given headers
// into every request. Returns nil if headers is empty (so the SDK uses its default).
func httpClientWithHeaders(headers map[string]string) *http.Client {
	if len(headers) == 0 {
		return nil
	}
	return &http.Client{
		Transport: &headerTransport{headers: headers},
	}
}

// headerTransport wraps http.DefaultTransport and adds custom headers.
type headerTransport struct {
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return http.DefaultTransport.RoundTrip(req)
}
