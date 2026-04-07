// Package plugins implements the plugin system (commands, hooks, agents).
// Compatible with the claude-code/plugins directory layout (plugin.json manifest).
package plugins

import (
	"context"
)

// Manifest describes a plugin, loaded from plugin.json.
type Manifest struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Commands    []CommandDef      `json:"commands,omitempty"`
	Hooks       []HookDef         `json:"hooks,omitempty"`
	Agents      []AgentDef        `json:"agents,omitempty"`
}

// CommandDef describes a plugin-provided command.
type CommandDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// HookDef describes a plugin-provided hook.
type HookDef struct {
	Event string `json:"event"`
	Cmd   string `json:"cmd"`
}

// AgentDef describes a plugin-provided agent.
type AgentDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Manager handles plugin discovery and lifecycle.
type Manager struct {
	plugins []*Manifest
}

// NewManager creates a new plugin manager.
func NewManager() *Manager {
	return &Manager{}
}

// Discover scans plugin directories and loads manifests.
func (m *Manager) Discover(ctx context.Context, dirs ...string) error {
	// TODO: implement plugin discovery
	return nil
}
