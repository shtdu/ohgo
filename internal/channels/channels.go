// Package channels implements IM channel integrations (Telegram, Slack, Discord, Feishu).
package channels

import (
	"context"
	"sync"
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
	mu       sync.RWMutex
	channels map[string]Channel
}

// NewRegistry creates a new channel registry.
func NewRegistry() *Registry {
	return &Registry{channels: make(map[string]Channel)}
}

// Register adds a channel integration.
func (r *Registry) Register(c Channel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.channels[c.Name()] = c
}

// Get retrieves a channel by name.
func (r *Registry) Get(name string) Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channels[name]
}

// List returns all registered channels.
func (r *Registry) List() []Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Channel, 0, len(r.channels))
	for _, c := range r.channels {
		out = append(out, c)
	}
	return out
}
