// Package channels implements IM channel integrations (Telegram, Slack, Discord, Feishu).
package channels

import (
	"context"
)

// Channel represents an IM channel integration.
type Channel interface {
	// Name returns the channel identifier (e.g. "telegram", "slack").
	Name() string

	// Connect establishes a connection to the IM service.
	Connect(ctx context.Context) error

	// Close shuts down the channel connection.
	Close() error
}

// Registry manages active channel integrations.
type Registry struct {
	channels map[string]Channel
}

// NewRegistry creates a new channel registry.
func NewRegistry() *Registry {
	return &Registry{channels: make(map[string]Channel)}
}

// Register adds a channel integration.
func (r *Registry) Register(c Channel) {
	r.channels[c.Name()] = c
}
