package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings()
	assert.Equal(t, "claude-sonnet-4-6", s.Model)
	assert.Equal(t, 16384, s.MaxTokens)
	assert.Equal(t, "anthropic", s.APIFormat)
	assert.Equal(t, "claude-api", s.ActiveProfile)
	assert.Equal(t, 200, s.MaxTurns)
	assert.Equal(t, "default", s.Permission.Mode)
	assert.True(t, s.Memory.Enabled)
}

func TestDefaultProviderProfiles(t *testing.T) {
	profiles := DefaultProviderProfiles()
	assert.Contains(t, profiles, "claude-api")
	assert.Contains(t, profiles, "openai-compatible")
	assert.Contains(t, profiles, "copilot")
	assert.Equal(t, "anthropic", profiles["claude-api"].Provider)
	assert.Equal(t, "openai", profiles["openai-compatible"].Provider)
}

func TestSettings_JSONRoundTrip(t *testing.T) {
	original := DefaultSettings()
	original.APIKey = "test-key"
	original.MaxTurns = 100

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Settings
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "test-key", decoded.APIKey)
	assert.Equal(t, 100, decoded.MaxTurns)
	assert.Equal(t, "claude-sonnet-4-6", decoded.Model)
}

func TestSettings_EmptyJSON(t *testing.T) {
	var s Settings
	err := json.Unmarshal([]byte("{}"), &s)
	require.NoError(t, err)
	// Zero values should be fine
	assert.Equal(t, "", s.Model)
	assert.Equal(t, 0, s.MaxTokens)
}

func TestSettings_MergedProfiles(t *testing.T) {
	s := DefaultSettings()
	// Add a custom profile
	s.Profiles["my-custom"] = ProviderProfile{
		Label:        "My Custom",
		Provider:     "openai",
		APIFormat:    "openai",
		AuthSource:   "openai_api_key",
		DefaultModel: "gpt-custom",
	}

	merged := s.MergedProfiles()
	assert.Contains(t, merged, "claude-api")       // built-in preserved
	assert.Contains(t, merged, "my-custom")         // custom added
	assert.Equal(t, "gpt-custom", merged["my-custom"].DefaultModel)
}

func TestSettings_ResolveProfile(t *testing.T) {
	s := DefaultSettings()
	name, profile := s.ResolveProfile("claude-api")
	assert.Equal(t, "claude-api", name)
	assert.Equal(t, "anthropic", profile.Provider)
}

func TestSettings_ResolveProfile_DefaultActive(t *testing.T) {
	s := DefaultSettings()
	name, _ := s.ResolveProfile("")
	assert.Equal(t, "claude-api", name)
}

func TestProviderProfile_ResolvedModel(t *testing.T) {
	tests := []struct {
		name     string
		profile  ProviderProfile
		expected string
	}{
		{"uses last model", ProviderProfile{LastModel: "claude-opus-4-6", DefaultModel: "claude-sonnet-4-6"}, "claude-opus-4-6"},
		{"falls back to default", ProviderProfile{DefaultModel: "claude-sonnet-4-6"}, "claude-sonnet-4-6"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.profile.ResolvedModel())
		})
	}
}

func TestPermissionSettings_JSONRoundTrip(t *testing.T) {
	original := PermissionSettings{
		Mode:           "plan",
		AllowedTools:   []string{"read", "glob"},
		DeniedTools:    []string{"bash"},
		PathRules:      []PathRuleConfig{{Pattern: "/tmp/*", Allow: true}},
		DeniedCommands: []string{"rm -rf"},
	}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded PermissionSettings
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, "plan", decoded.Mode)
	require.Len(t, decoded.AllowedTools, 2)
	require.Len(t, decoded.PathRules, 1)
	assert.Equal(t, "/tmp/*", decoded.PathRules[0].Pattern)
}
