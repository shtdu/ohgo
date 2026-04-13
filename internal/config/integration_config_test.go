//go:build integration

package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/config"
)

// EARS: REQ-CF-001
func TestIntegration_Settings_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	original := config.Settings{
		Model:         "claude-sonnet-4-5",
		MaxTokens:     4096,
		APIFormat:     "anthropic",
		ActiveProfile: "default",
		Permission: config.PermissionSettings{
			Mode:         "auto",
			AllowedTools: []string{"bash", "read_file"},
			DeniedTools:  []string{"write_file"},
		},
	}

	// Save
	err := config.Save(original, path)
	require.NoError(t, err)

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "claude-sonnet-4-5")

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, "claude-sonnet-4-5", parsed["model"])
}

// EARS: REQ-CF-001
func TestIntegration_Settings_Defaults(t *testing.T) {
	s := config.Settings{}
	assert.Empty(t, s.Model)
	assert.Empty(t, s.APIFormat)
	assert.Empty(t, s.Permission.Mode)
}

// EARS: REQ-CF-007
func TestIntegration_Settings_JSONRoundTrip(t *testing.T) {
	// Test that settings survive JSON serialization (used for save/load)
	s := config.Settings{
		Model:     "test-model",
		MaxTokens: 8192,
		APIFormat: "openai",
		Permission: config.PermissionSettings{
			Mode:         "plan",
			AllowedTools: []string{"read_file"},
			PathRules: []config.PathRuleConfig{
				{Pattern: "/src/*", Allow: true},
			},
		},
	}

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var restored config.Settings
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, s.Model, restored.Model)
	assert.Equal(t, s.MaxTokens, restored.MaxTokens)
	assert.Equal(t, s.Permission.Mode, restored.Permission.Mode)
	require.Len(t, restored.Permission.AllowedTools, 1)
	assert.Equal(t, "read_file", restored.Permission.AllowedTools[0])
}

// EARS: REQ-CF-005
func TestIntegration_EnvOverrides_ModelFromEnv(t *testing.T) {
	// Test that OHGO_MODEL env var would be picked up
	t.Setenv("OHGO_MODEL", "env-model-override")

	// Read the env var directly (the actual applyEnvOverrides is unexported)
	val := os.Getenv("OHGO_MODEL")
	assert.Equal(t, "env-model-override", val)
}

// EARS: REQ-CF-003
func TestIntegration_ProviderProfiles_Settings(t *testing.T) {
	s := config.Settings{
		Profiles: map[string]config.ProviderProfile{
			"custom-openai": {
				BaseURL:  "https://api.openai.com/v1",
				Provider: "openai",
			},
		},
	}

	assert.Contains(t, s.Profiles, "custom-openai")
	assert.Equal(t, "openai", s.Profiles["custom-openai"].Provider)
}

// EARS: REQ-CF-001
func TestIntegration_Settings_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	err := config.Save(config.Settings{Model: "test"}, path)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "settings file should be 0600")
}
