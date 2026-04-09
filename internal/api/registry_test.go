package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/config"
)

func TestRegistry_AnthropicFactory(t *testing.T) {
	r := NewRegistry()
	cfg := &config.Settings{
		APIKey: "sk-test-123",
	}

	client, err := r.CreateClient(cfg, "")
	require.NoError(t, err)
	require.NotNil(t, client)

	// Should be an AnthropicClient.
	_, ok := client.(*AnthropicClient)
	assert.True(t, ok, "expected *AnthropicClient")
}

func TestRegistry_CustomBaseURL(t *testing.T) {
	r := NewRegistry()
	cfg := &config.Settings{
		APIKey: "sk-test",
		Profiles: map[string]config.ProviderProfile{
			"custom": {
				Label:     "Custom",
				Provider:  "anthropic",
				APIFormat: "anthropic",
				BaseURL:   "https://custom-api.example.com/v1/messages",
			},
		},
		ActiveProfile: "custom",
	}

	client, err := r.CreateClient(cfg, "custom")
	require.NoError(t, err)
	ac, ok := client.(*AnthropicClient)
	require.True(t, ok)
	assert.Equal(t, "https://custom-api.example.com/v1/messages", ac.baseURL)
}

func TestRegistry_EnvVarFallback(t *testing.T) {
	r := NewRegistry()
	cfg := &config.Settings{
		ActiveProfile: "openai-compatible",
		Profiles:      config.DefaultProviderProfiles(),
	}

	// Without env var, the openai factory won't be registered yet.
	// Test with anthropic format using env var.
	t.Setenv("ANTHROPIC_API_KEY", "sk-env-key")
	cfg.APIKey = "" // clear direct key
	cfg.ActiveProfile = "claude-api"

	client, err := r.CreateClient(cfg, "")
	require.NoError(t, err)
	ac, ok := client.(*AnthropicClient)
	require.True(t, ok)
	assert.Equal(t, "sk-env-key", ac.apiKey)
}

func TestRegistry_UnsupportedFormat(t *testing.T) {
	r := NewRegistry()
	cfg := &config.Settings{
		ActiveProfile: "unknown",
		Profiles: map[string]config.ProviderProfile{
			"unknown": {
				APIFormat: "unknown_format",
			},
		},
	}

	_, err := r.CreateClient(cfg, "unknown")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported api_format")
}

func TestRegistry_CustomFactory(t *testing.T) {
	r := NewRegistry()
	called := false
	r.Register("custom", func(profile config.ProviderProfile, apiKey string) (Client, error) {
		called = true
		assert.Equal(t, "test-key", apiKey)
		// Return a minimal mock that satisfies Client.
		return &AnthropicClient{}, nil
	})

	cfg := &config.Settings{
		APIKey: "test-key",
		Profiles: map[string]config.ProviderProfile{
			"my-profile": {
				APIFormat: "custom",
			},
		},
		ActiveProfile: "my-profile",
	}

	client, err := r.CreateClient(cfg, "")
	require.NoError(t, err)
	assert.True(t, called)
	assert.NotNil(t, client)
}

func TestRegistry_ResolvesProfileFromFlag(t *testing.T) {
	r := NewRegistry()
	cfg := &config.Settings{
		APIKey: "sk-test",
		Profiles: map[string]config.ProviderProfile{
			"alt": {
				Label:        "Alt Anthropic",
				Provider:     "anthropic",
				APIFormat:    "anthropic",
				DefaultModel: "claude-haiku-4-5-20251001",
			},
		},
	}

	client, err := r.CreateClient(cfg, "alt")
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestEnvKey(t *testing.T) {
	tests := []struct {
		authSource string
		want       string
	}{
		{"anthropic_api_key", "ANTHROPIC_API_KEY"},
		{"openai_api_key", "OPENAI_API_KEY"},
		{"copilot_oauth", "GITHUB_TOKEN"},
		{"codex_subscription", "CODEX_API_KEY"},
		{"unknown", ""},
	}
	for _, tt := range tests {
		t.Run(tt.authSource, func(t *testing.T) {
			p := config.ProviderProfile{AuthSource: tt.authSource}
			assert.Equal(t, tt.want, p.EnvKey())
		})
	}
}

func TestRegistry_DefaultProfile(t *testing.T) {
	r := NewRegistry()
	cfg := &config.Settings{
		APIKey: "sk-test",
	}
	// Empty profile name should resolve to claude-api.
	client, err := r.CreateClient(cfg, "")
	require.NoError(t, err)
	_, ok := client.(*AnthropicClient)
	assert.True(t, ok)
}
