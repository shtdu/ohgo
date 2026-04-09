// Package bridge implements subscription bridges (Claude CLI, Codex CLI).
package bridge

import "context"

// Bridge connects to an existing subscription service to proxy API calls.
type Bridge interface {
	// Name returns the bridge identifier (e.g. "claude", "codex").
	Name() string

	// Connect establishes a connection to the subscription service.
	Connect(ctx context.Context) error

	// Close shuts down the bridge connection.
	Close() error
}
