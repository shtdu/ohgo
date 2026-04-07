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
