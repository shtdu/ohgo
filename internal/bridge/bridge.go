// Package bridge implements subscription bridges (Claude CLI, Codex CLI).
package bridge

import (
	"context"
)

// Bridge connects to an existing subscription service to proxy API calls.
type Bridge interface {
	// Name returns the bridge identifier (e.g. "claude", "codex").
	Name() string

	// Connect establishes a connection to the subscription service.
	Connect(ctx context.Context) error

	// Close shuts down the bridge connection.
	Close() error
}

// ClaudeCLI bridges to the Claude CLI subscription.
type ClaudeCLI struct{}

// Name returns the bridge name.
func (c *ClaudeCLI) Name() string { return "claude" }

// Connect establishes a bridge to the Claude CLI.
func (c *ClaudeCLI) Connect(ctx context.Context) error { return nil }

// Close shuts down the bridge.
func (c *ClaudeCLI) Close() error { return nil }
