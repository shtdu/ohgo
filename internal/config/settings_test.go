package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultsOnly(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENHARNESS_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")

	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "claude-sonnet-4-6", s.Model)
	assert.Equal(t, 16384, s.MaxTokens)
	assert.Equal(t, "claude-api", s.ActiveProfile)
}

func TestLoad_FromFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENHARNESS_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")

	configData := `{"model": "claude-opus-4-6", "max_tokens": 8192, "max_turns": 50}`
	err := os.WriteFile(filepath.Join(tmp, "settings.json"), []byte(configData), 0o644)
	require.NoError(t, err)

	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "claude-opus-4-6", s.Model)
	assert.Equal(t, 8192, s.MaxTokens)
	assert.Equal(t, 50, s.MaxTurns)
}

func TestLoad_EnvOverrides(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("OPENHARNESS_MODEL", "gpt-4")
	t.Setenv("OPENHARNESS_MAX_TOKENS", "4096")
	t.Setenv("ANTHROPIC_API_KEY", "sk-test-123")

	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "gpt-4", s.Model)
	assert.Equal(t, 4096, s.MaxTokens)
	assert.Equal(t, "sk-test-123", s.APIKey)
}

func TestLoad_MalformedJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("ANTHROPIC_API_KEY", "")

	err := os.WriteFile(filepath.Join(tmp, "settings.json"), []byte("{invalid"), 0o644)
	require.NoError(t, err)

	// loadFromFile should error
	_, err = loadFromFile(filepath.Join(tmp, "settings.json"))
	assert.Error(t, err)

	// Manager.Load should still return defaults (ignoring malformed file)
	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "claude-sonnet-4-6", s.Model, "should fall back to defaults")
}

func TestSaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "settings.json")

	original := DefaultSettings()
	original.APIKey = "test-key"
	original.MaxTurns = 42

	err := Save(original, path)
	require.NoError(t, err)

	loaded, err := loadFromFile(path)
	require.NoError(t, err)
	assert.Equal(t, "test-key", loaded.APIKey)
	assert.Equal(t, 42, loaded.MaxTurns)
}

func TestMergeSettings(t *testing.T) {
	base := DefaultSettings()
	override := Settings{
		Model:     "gpt-4",
		MaxTokens: 4096,
		Provider:  "openai",
	}

	result := mergeSettings(base, override)
	assert.Equal(t, "gpt-4", result.Model)
	assert.Equal(t, 4096, result.MaxTokens)
	assert.Equal(t, "openai", result.Provider)
	// Non-overridden fields should keep base values
	assert.Equal(t, base.MaxTurns, result.MaxTurns)
	assert.Equal(t, base.ActiveProfile, result.ActiveProfile)
}

func TestParseBoolEnv(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"1", true},
		{"True", true},
		{"false", false},
		{"0", false},
		{"", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, parseBoolEnv(tt.input))
	}
}

// --- Environment Override Tests (via Manager.Load) ---

func TestLoad_AnthropicBaseURL(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENHARNESS_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")
	t.Setenv("ANTHROPIC_BASE_URL", "https://custom.api.com")

	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://custom.api.com", s.BaseURL)
}

func TestLoad_OpenHarnessBaseURL(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENHARNESS_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")
	t.Setenv("ANTHROPIC_BASE_URL", "")
	t.Setenv("OPENHARNESS_BASE_URL", "https://oh.api.com")

	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://oh.api.com", s.BaseURL)
}

func TestLoad_AnthropicBaseURLPreferred(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENHARNESS_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")
	t.Setenv("ANTHROPIC_BASE_URL", "https://anthropic.api.com")
	t.Setenv("OPENHARNESS_BASE_URL", "https://oh.api.com")

	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "https://anthropic.api.com", s.BaseURL)
}

func TestLoad_APIFormat(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENHARNESS_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")
	t.Setenv("OPENHARNESS_API_FORMAT", "openai")

	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "openai", s.APIFormat)
}

func TestLoad_Provider(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENHARNESS_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")
	t.Setenv("OPENHARNESS_PROVIDER", "openai")

	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "openai", s.Provider)
}

func TestLoad_MaxTurns(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENHARNESS_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")
	t.Setenv("OPENHARNESS_MAX_TURNS", "100")

	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 100, s.MaxTurns)
}

func TestLoad_InvalidMaxTokens(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENHARNESS_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")
	t.Setenv("OPENHARNESS_MAX_TOKENS", "notanumber")

	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 16384, s.MaxTokens, "should stay at default when env value is invalid")
}

func TestLoad_AnthropicModelPreferred(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("ANTHROPIC_MODEL", "claude-opus-4-6")
	t.Setenv("OPENHARNESS_MODEL", "gpt-4")

	mgr := NewManager(tmp)
	s, err := mgr.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "claude-opus-4-6", s.Model)
}

// --- Merge Settings Tests ---

func TestMergeSettings_Profiles(t *testing.T) {
	base := DefaultSettings()
	base.Profiles = nil

	override := Settings{
		Profiles: map[string]ProviderProfile{
			"custom": {
				Label:        "Custom Profile",
				Provider:     "custom",
				APIFormat:    "openai",
				DefaultModel: "custom-model",
			},
		},
	}

	result := mergeSettings(base, override)
	require.NotNil(t, result.Profiles)
	assert.Contains(t, result.Profiles, "custom")
	assert.Equal(t, "Custom Profile", result.Profiles["custom"].Label)
}

func TestMergeSettings_Permission(t *testing.T) {
	base := DefaultSettings()

	override := Settings{
		Permission: PermissionSettings{
			Mode:        "auto",
			DeniedTools: []string{"rm"},
		},
	}

	result := mergeSettings(base, override)
	assert.Equal(t, "auto", result.Permission.Mode)
	assert.Equal(t, []string{"rm"}, result.Permission.DeniedTools)
}

func TestMergeSettings_VimMode(t *testing.T) {
	base := DefaultSettings()

	override := Settings{
		VimMode: true,
	}

	result := mergeSettings(base, override)
	assert.True(t, result.VimMode)
}

func TestMergeSettings_Verbose(t *testing.T) {
	base := DefaultSettings()

	override := Settings{
		Verbose: true,
	}

	result := mergeSettings(base, override)
	assert.True(t, result.Verbose)
}

func TestMergeSettings_EmptyOverride(t *testing.T) {
	base := DefaultSettings()
	override := Settings{}

	result := mergeSettings(base, override)
	assert.Equal(t, base.Model, result.Model)
	assert.Equal(t, base.MaxTokens, result.MaxTokens)
	assert.Equal(t, base.MaxTurns, result.MaxTurns)
	assert.Equal(t, base.ActiveProfile, result.ActiveProfile)
	assert.Equal(t, base.APIFormat, result.APIFormat)
	assert.Equal(t, base.Theme, result.Theme)
	assert.Equal(t, base.OutputStyle, result.OutputStyle)
	assert.False(t, result.VimMode)
	assert.False(t, result.Verbose)
}

func TestMergeSettings_BaseURL(t *testing.T) {
	base := DefaultSettings()

	override := Settings{
		BaseURL: "https://custom.api.com",
	}

	result := mergeSettings(base, override)
	assert.Equal(t, "https://custom.api.com", result.BaseURL)
}

func TestMergeSettings_ActiveProfile(t *testing.T) {
	base := DefaultSettings()

	override := Settings{
		ActiveProfile: "openai-compatible",
	}

	result := mergeSettings(base, override)
	assert.Equal(t, "openai-compatible", result.ActiveProfile)
}

func TestMergeSettings_SystemPrompt(t *testing.T) {
	base := DefaultSettings()

	override := Settings{
		SystemPrompt: "You are a helpful assistant.",
	}

	result := mergeSettings(base, override)
	assert.Equal(t, "You are a helpful assistant.", result.SystemPrompt)
}

func TestMergeSettings_Theme(t *testing.T) {
	base := DefaultSettings()

	override := Settings{
		Theme: "dark",
	}

	result := mergeSettings(base, override)
	assert.Equal(t, "dark", result.Theme)
}

func TestMergeSettings_OutputStyle(t *testing.T) {
	base := DefaultSettings()

	override := Settings{
		OutputStyle: "json",
	}

	result := mergeSettings(base, override)
	assert.Equal(t, "json", result.OutputStyle)
}

// --- Save Error Tests ---

func TestSave_InvalidPath(t *testing.T) {
	s := DefaultSettings()
	err := Save(s, "/nonexistent/deeply/nested/dir/settings.json")
	assert.Error(t, err)
}

func TestLoadConfig_WithConfigDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENHARNESS_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")

	// Write settings to the custom config dir
	configData := `{"model": "custom-model", "max_tokens": 9999}`
	err := os.WriteFile(filepath.Join(tmp, "settings.json"), []byte(configData), 0o644)
	require.NoError(t, err)

	s, err := loadConfig(context.Background(), tmp)
	require.NoError(t, err)
	assert.Equal(t, "custom-model", s.Model)
	assert.Equal(t, 9999, s.MaxTokens)
}

func TestLoadConfig_EmptyConfigDir(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENHARNESS_MODEL", "")
	t.Setenv("ANTHROPIC_MODEL", "")
	// Empty configDir should use default path via ConfigFilePath
	s, err := loadConfig(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "claude-sonnet-4-6", s.Model)
}
