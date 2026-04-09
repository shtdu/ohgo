package api

import (
	"fmt"
	"os"

	"github.com/shtdu/ohgo/internal/config"
)

// ClientFactory creates a Client from a resolved profile and API key.
type ClientFactory func(profile config.ProviderProfile, apiKey string) (Client, error)

// Registry maps APIFormat strings to ClientFactory functions.
type Registry struct {
	factories map[string]ClientFactory
}

// NewRegistry creates a registry pre-loaded with the built-in Anthropic provider.
func NewRegistry() *Registry {
	r := &Registry{
		factories: make(map[string]ClientFactory),
	}
	// Register built-in Anthropic provider.
	r.Register("anthropic", func(profile config.ProviderProfile, apiKey string) (Client, error) {
		opts := []AnthropicOption{WithAPIKey(apiKey)}
		if profile.BaseURL != "" {
			opts = append(opts, WithBaseURL(profile.BaseURL))
		}
		return NewAnthropicClient(opts...), nil
	})

	// Register built-in OpenAI-compatible provider.
	r.Register("openai", newOpenAIFactory)

	// Register built-in Copilot provider.
	r.Register("copilot", newCopilotFactory)
	return r
}

// Register adds or replaces a factory for the given apiFormat.
func (r *Registry) Register(apiFormat string, factory ClientFactory) {
	r.factories[apiFormat] = factory
}

// CreateClient resolves the active profile from settings and constructs
// the appropriate Client.
func (r *Registry) CreateClient(cfg *config.Settings, profileName string) (Client, error) {
	_, profile := cfg.ResolveProfile(profileName)

	// Settings-level overrides take precedence over profile values.
	if cfg.BaseURL != "" {
		profile.BaseURL = cfg.BaseURL
	}

	// Resolve API key: settings → profile env var fallback.
	apiKey := cfg.ResolveAPIKey()
	if apiKey == "" {
		if envKey := profile.EnvKey(); envKey != "" {
			apiKey = os.Getenv(envKey)
		}
	}

	factory, ok := r.factories[profile.APIFormat]
	if !ok {
		return nil, fmt.Errorf("unsupported api_format %q for profile %q", profile.APIFormat, profileName)
	}
	return factory(profile, apiKey)
}
