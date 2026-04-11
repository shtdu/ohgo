// Package config implements multi-layer configuration and profile management.
package config

import (
	"context"
)

// Manager handles multi-layer config loading.
type Manager struct {
	configDir string
}

// NewManager creates a config manager for the given config directory.
// If configDir is empty, uses the default location.
func NewManager(configDir string) *Manager {
	return &Manager{configDir: configDir}
}

// ConfigDir returns the config directory for this manager.
func (m *Manager) ConfigDir() string {
	if m.configDir != "" {
		return m.configDir
	}
	dir, err := ConfigDir()
	if err != nil {
		return ""
	}
	return dir
}

// Load reads and merges config from all layers.
// Returns settings with the following precedence (highest first):
// env vars > project config > user config > defaults.
func (m *Manager) Load(ctx context.Context) (*Settings, error) {
	return loadConfig(ctx, m.configDir)
}

// loadConfig is the shared implementation used by Manager.Load.
func loadConfig(_ context.Context, configDir string) (*Settings, error) {
	s := DefaultSettings()

	// User config
	var userPath string
	if configDir != "" {
		userPath = configDir + "/settings.json"
	} else {
		var err error
		userPath, err = ConfigFilePath()
		if err != nil {
			return nil, err
		}
	}
	if fileSettings, err := loadFromFile(userPath); err == nil {
		s = mergeSettings(s, fileSettings)
	}

	// Project config
	if cwd, err := WorkingDir(); err == nil {
		projectPath := cwd + "/.openharness/settings.json"
		if projSettings, err := loadFromFile(projectPath); err == nil {
			s = mergeSettings(s, projSettings)
		}
	}

	// Environment overrides
	applyEnvOverrides(&s)

	return &s, nil
}
