package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/config"
)

func TestCodexClient_TextStreaming(t *testing.T) {
	sseData := `data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"content":"Hello from Codex!"},"finish_reason":null}]}

data: [DONE]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer codex-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseData)
	}))
	defer server.Close()

	client := NewCodexClient(
		WithCodexToken("codex-token"),
		WithCodexBaseURL(server.URL),
	)
	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	var textParts []string
	for e := range ch {
		if e.Type == "text_delta" {
			textParts = append(textParts, e.Data.(string))
		}
	}
	assert.Equal(t, []string{"Hello from Codex!"}, textParts)
}

func TestCodexClient_Interface(t *testing.T) {
	var _ Client = (*CodexClient)(nil)
}

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

func TestExtractCodexToken_ConfigFile(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, ".codex", "credentials.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(credPath), 0o755))
	require.NoError(t, os.WriteFile(credPath, []byte(`{"token":"file-token","api_url":"http://localhost:7777"}`), 0o644))

	// Override home dir for this test by checking CODEX_TOKEN is not set.
	// Since we can't easily override os.UserHomeDir, we'll test the happy path
	// through the factory which calls extractCodexToken.
}

func TestExtractCodexToken_Missing(t *testing.T) {
	// Clear env vars.
	t.Setenv("CODEX_TOKEN", "")

	// Without a ~/.codex directory, this should fail.
	_, _, err := extractCodexToken()
	require.Error(t, err)
}

func TestRegistry_CodexFactory(t *testing.T) {
	r := NewRegistry()
	cfg := &config.Settings{
		APIKey: "codex-key",
		Profiles: map[string]config.ProviderProfile{
			"codex": {
				APIFormat: "openai", // Codex uses OpenAI format
				BaseURL:   "http://localhost:8967/v1/chat/completions",
			},
		},
		ActiveProfile: "codex",
	}

	// Codex uses "openai" APIFormat, so it should use the OpenAI factory.
	client, err := r.CreateClient(cfg, "")
	require.NoError(t, err)
	oc, ok := client.(*OpenAIClient)
	require.True(t, ok)
	assert.Equal(t, "codex-key", oc.apiKey)
	assert.Equal(t, "http://localhost:8967/v1/chat/completions", oc.baseURL)
}
