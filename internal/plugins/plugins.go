// Package plugins implements the plugin system (commands, hooks, agents).
// Compatible with the claude-code/plugins directory layout (plugin.json manifest).
package plugins

import (
	"context"
	"sync"
)

// Manager handles plugin discovery and lifecycle.
type Manager struct {
	mu      sync.RWMutex
	plugins []*LoadedPlugin
}

// NewManager creates a new plugin manager.
func NewManager() *Manager {
	return &Manager{}
}

// Discover scans plugin directories and loads manifests.
// Replaces any previously loaded plugins with the newly discovered set.
func (m *Manager) Discover(ctx context.Context, dirs ...string) error {
	plugins, err := Discover(ctx, dirs)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.plugins = plugins
	m.mu.Unlock()

	return nil
}

// List returns all loaded plugins sorted by name.
func (m *Manager) List() []*LoadedPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*LoadedPlugin, len(m.plugins))
	copy(result, m.plugins)
	return result
}

// Get returns a plugin by name, or nil if not found.
func (m *Manager) Get(name string) *LoadedPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.plugins {
		if p.Manifest.Name == name {
			return p
		}
	}
	return nil
}
