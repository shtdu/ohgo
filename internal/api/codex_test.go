package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/config"
)

func TestExtractCodexToken_EnvVar(t *testing.T) {
	t.Setenv("CODEX_TOKEN", "env-token")
	t.Setenv("CODEX_API_URL", "http://localhost:9999/v1")

	token, baseURL, err := extractCodexToken()
	require.NoError(t, err)
	assert.Equal(t, "env-token", token)
	assert.Equal(t, "http://localhost:9999/v1", baseURL)
}

func TestExtractCodexToken_EnvVarDefaultURL(t *testing.T) {
	t.Setenv("CODEX_TOKEN", "env-token")

	token, baseURL, err := extractCodexToken()
	require.NoError(t, err)
	assert.Equal(t, "env-token", token)
	assert.Equal(t, defaultCodexBaseURL, baseURL)
}

func TestExtractCodexToken_Missing(t *testing.T) {
	t.Setenv("CODEX_TOKEN", "")

	_, _, err := extractCodexToken()
	require.Error(t, err)
}

func TestRegistry_CodexProfileUsesOpenAIFactory(t *testing.T) {
	r := NewRegistry()
	cfg := &config.Settings{
		APIKey: "codex-key",
		Profiles: map[string]config.ProviderProfile{
			"codex": {
				APIFormat:  "openai",
				BaseURL:    "http://localhost:8967/v1/chat/completions",
				AuthSource: "codex_subscription",
			},
		},
		ActiveProfile: "codex",
	}

	client, err := r.CreateClient(cfg, "")
	require.NoError(t, err)
	oc, ok := client.(*OpenAIClient)
	require.True(t, ok)
	assert.Equal(t, "codex-key", oc.apiKey)
	assert.Equal(t, "http://localhost:8967/v1/chat/completions", oc.baseURL)
}

func TestExtractCodexToken_ConfigFileParsing(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, ".codex", "credentials.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(credPath), 0o755))
	require.NoError(t, os.WriteFile(credPath, []byte(`{"token":"file-token","api_url":"http://localhost:7777"}`), 0o644))

	// Verify JSON structure matches what extractCodexToken expects.
	data, err := os.ReadFile(credPath)
	require.NoError(t, err)

	var credData map[string]any
	require.NoError(t, json.Unmarshal(data, &credData))
	assert.Equal(t, "file-token", credData["token"])
	assert.Equal(t, "http://localhost:7777", credData["api_url"])
}

func TestExtractCodexToken_ConfigFileWithAPIKey(t *testing.T) {
	// Set HOME to a temp dir so the function reads our crafted file.
	t.Setenv("CODEX_TOKEN", "") // clear env to force file read
	dir := t.TempDir()
	credDir := filepath.Join(dir, ".codex")
	credPath := filepath.Join(credDir, "credentials.json")
	require.NoError(t, os.MkdirAll(credDir, 0o755))
	require.NoError(t, os.WriteFile(credPath, []byte(`{"api_key":"ak-from-file"}`), 0o644))
	t.Setenv("HOME", dir)

	token, _, err := extractCodexToken()
	require.NoError(t, err)
	assert.Equal(t, "ak-from-file", token)
}

func TestExtractCodexToken_ConfigFileNoToken(t *testing.T) {
	t.Setenv("CODEX_TOKEN", "")
	dir := t.TempDir()
	credDir := filepath.Join(dir, ".codex")
	credPath := filepath.Join(credDir, "credentials.json")
	require.NoError(t, os.MkdirAll(credDir, 0o755))
	require.NoError(t, os.WriteFile(credPath, []byte(`{"other":"data"}`), 0o644))
	t.Setenv("HOME", dir)

	_, _, err := extractCodexToken()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no token found")
}

func TestExtractCodexToken_ConfigFileInvalidJSON(t *testing.T) {
	t.Setenv("CODEX_TOKEN", "")
	dir := t.TempDir()
	credDir := filepath.Join(dir, ".codex")
	credPath := filepath.Join(credDir, "credentials.json")
	require.NoError(t, os.MkdirAll(credDir, 0o755))
	require.NoError(t, os.WriteFile(credPath, []byte(`not json`), 0o644))
	t.Setenv("HOME", dir)

	_, _, err := extractCodexToken()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse credentials")
}

func TestExtractCodexToken_ConfigFileCustomURL(t *testing.T) {
	t.Setenv("CODEX_TOKEN", "")
	dir := t.TempDir()
	credDir := filepath.Join(dir, ".codex")
	credPath := filepath.Join(credDir, "credentials.json")
	require.NoError(t, os.MkdirAll(credDir, 0o755))
	require.NoError(t, os.WriteFile(credPath, []byte(`{"token":"t1","api_url":"http://custom:1234"}`), 0o644))
	t.Setenv("HOME", dir)

	token, baseURL, err := extractCodexToken()
	require.NoError(t, err)
	assert.Equal(t, "t1", token)
	assert.Equal(t, "http://custom:1234", baseURL)
}

func TestNewOpenAIFactory_WithBaseURL(t *testing.T) {
	profile := config.ProviderProfile{
		APIFormat: "openai",
		BaseURL:   "http://localhost:8080/v1",
	}
	client, err := newOpenAIFactory(profile, "test-key")
	require.NoError(t, err)
	oc, ok := client.(*OpenAIClient)
	require.True(t, ok)
	assert.Equal(t, "test-key", oc.apiKey)
	assert.Equal(t, "http://localhost:8080/v1", oc.baseURL)
}

func TestNewOpenAIFactory_WithoutBaseURL(t *testing.T) {
	profile := config.ProviderProfile{
		APIFormat: "openai",
	}
	client, err := newOpenAIFactory(profile, "key")
	require.NoError(t, err)
	oc, ok := client.(*OpenAIClient)
	require.True(t, ok)
	assert.Equal(t, defaultOpenAIBaseURL, oc.baseURL)
}

func TestNewOpenAIFactory_CodexWithEnvToken(t *testing.T) {
	t.Setenv("CODEX_TOKEN", "codex-env-token")
	t.Setenv("CODEX_API_URL", "http://codex:9999")
	profile := config.ProviderProfile{
		APIFormat:  "openai",
		AuthSource: "codex_subscription",
	}
	client, err := newOpenAIFactory(profile, "")
	require.NoError(t, err)
	oc, ok := client.(*OpenAIClient)
	require.True(t, ok)
	assert.Equal(t, "codex-env-token", oc.apiKey)
	assert.Equal(t, "http://codex:9999", oc.baseURL)
}

func TestNewOpenAIFactory_CodexNoToken(t *testing.T) {
	t.Setenv("CODEX_TOKEN", "")
	t.Setenv("HOME", "/nonexistent/path/that/does/not/exist")
	profile := config.ProviderProfile{
		APIFormat:  "openai",
		AuthSource: "codex_subscription",
	}
	_, err := newOpenAIFactory(profile, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "codex")
}
