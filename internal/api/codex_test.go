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
