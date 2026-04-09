package bridge

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
)

// Session represents an active bridge session.
type Session struct {
	ID        string
	Bridge    Bridge
	connected atomic.Bool
}

// SessionRunner manages the lifecycle of bridge sessions.
type SessionRunner struct {
	manager  *Manager
	sessions map[string]*Session
	mu       sync.Mutex
}

// NewSessionRunner creates a new session runner.
func NewSessionRunner(m *Manager) *SessionRunner {
	return &SessionRunner{
		manager:  m,
		sessions: make(map[string]*Session),
	}
}

// Start creates and starts a new bridge session.
func (r *SessionRunner) Start(ctx context.Context, bridgeName string) (*Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.manager.Get(bridgeName)
	if !ok {
		return nil, fmt.Errorf("bridge %q not found", bridgeName)
	}

	if err := b.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect bridge %q: %w", bridgeName, err)
	}

	sessionID := bridgeName + "-session"
	session := &Session{
		ID:     sessionID,
		Bridge: b,
	}
	session.connected.Store(true)
	r.sessions[sessionID] = session

	return session, nil
}

// Stop terminates a bridge session.
func (r *SessionRunner) Stop(sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session %q not found", sessionID)
	}

	if err := session.Bridge.Close(); err != nil {
		return err
	}
	session.connected.Store(false)
	delete(r.sessions, sessionID)

	return nil
}

// ClaudeCLI bridges to the Claude CLI subscription.
type ClaudeCLI struct {
	binaryPath string
	connected  bool
	mu         sync.Mutex
}

// NewClaudeCLI creates a new Claude CLI bridge.
func NewClaudeCLI() *ClaudeCLI {
	return &ClaudeCLI{}
}

// Name returns the bridge name.
func (c *ClaudeCLI) Name() string { return "claude" }

// Connect searches for the claude binary and verifies connectivity.
func (c *ClaudeCLI) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude CLI not found in PATH")
	}
	c.binaryPath = path
	c.connected = true
	return nil
}

// Close shuts down the bridge.
func (c *ClaudeCLI) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
	return nil
}

// IsConnected returns whether the bridge is connected.
func (c *ClaudeCLI) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// CodexBridge bridges to the Codex CLI subscription.
type CodexBridge struct {
	binaryPath string
	connected  bool
	mu         sync.Mutex
}

// NewCodexBridge creates a new Codex CLI bridge.
func NewCodexBridge() *CodexBridge {
	return &CodexBridge{}
}

// Name returns the bridge name.
func (c *CodexBridge) Name() string { return "codex" }

// Connect searches for the codex binary and verifies connectivity.
func (c *CodexBridge) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path, err := exec.LookPath("codex")
	if err != nil {
		return fmt.Errorf("codex CLI not found in PATH")
	}
	c.binaryPath = path
	c.connected = true
	return nil
}

// Close shuts down the bridge.
func (c *CodexBridge) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
	return nil
}

// IsConnected returns whether the bridge is connected.
func (c *CodexBridge) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}
