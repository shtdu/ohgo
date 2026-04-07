package config

import "encoding/json"

// PathRuleConfig is a glob-pattern path permission rule.
type PathRuleConfig struct {
	Pattern string `json:"pattern"`
	Allow   bool   `json:"allow"`
}

// PermissionSettings configures the permission mode.
type PermissionSettings struct {
	Mode           string           `json:"mode"`
	AllowedTools   []string         `json:"allowed_tools"`
	DeniedTools    []string         `json:"denied_tools"`
	PathRules      []PathRuleConfig `json:"path_rules"`
	DeniedCommands []string         `json:"denied_commands"`
}

// MemorySettings configures the memory system.
type MemorySettings struct {
	Enabled            bool `json:"enabled"`
	MaxFiles           int  `json:"max_files"`
	MaxEntrypointLines int  `json:"max_entrypoint_lines"`
}

// ProviderProfile is a named provider workflow configuration.
type ProviderProfile struct {
	Label         string   `json:"label"`
	Provider      string   `json:"provider"`
	APIFormat     string   `json:"api_format"`
	AuthSource    string   `json:"auth_source"`
	DefaultModel  string   `json:"default_model"`
	BaseURL       string   `json:"base_url,omitempty"`
	LastModel     string   `json:"last_model,omitempty"`
	CredentialSlot string  `json:"credential_slot,omitempty"`
	AllowedModels []string `json:"allowed_models,omitempty"`
}

// ResolvedAuth holds normalized auth material for constructing API clients.
type ResolvedAuth struct {
	Provider string
	AuthKind string
	Value    string
	Source   string
	State    string
}

// Settings is the main configuration model.
type Settings struct {
	// API configuration
	APIKey        string                     `json:"api_key,omitempty"`
	Model         string                     `json:"model"`
	MaxTokens     int                        `json:"max_tokens"`
	BaseURL       string                     `json:"base_url,omitempty"`
	APIFormat     string                     `json:"api_format"`
	Provider      string                     `json:"provider"`
	ActiveProfile string                     `json:"active_profile"`
	Profiles      map[string]ProviderProfile `json:"profiles"`
	MaxTurns      int                        `json:"max_turns"`

	// Behavior
	SystemPrompt string             `json:"system_prompt,omitempty"`
	Permission   PermissionSettings `json:"permission"`
	Memory       MemorySettings     `json:"memory"`

	// UI
	Theme       string `json:"theme"`
	OutputStyle string `json:"output_style"`
	VimMode     bool   `json:"vim_mode"`
	Verbose     bool   `json:"verbose"`
}

// DefaultSettings returns settings with sensible defaults.
func DefaultSettings() Settings {
	return Settings{
		Model:         "claude-sonnet-4-6",
		MaxTokens:     16384,
		APIFormat:     "anthropic",
		ActiveProfile: "claude-api",
		Profiles:      DefaultProviderProfiles(),
		MaxTurns:      200,
		Permission: PermissionSettings{
			Mode: "default",
		},
		Memory: MemorySettings{
			Enabled:            true,
			MaxFiles:           5,
			MaxEntrypointLines: 200,
		},
		Theme:       "default",
		OutputStyle: "default",
	}
}

// DefaultProviderProfiles returns the built-in provider catalog.
func DefaultProviderProfiles() map[string]ProviderProfile {
	return map[string]ProviderProfile{
		"claude-api": {
			Label:        "Anthropic-Compatible API",
			Provider:     "anthropic",
			APIFormat:    "anthropic",
			AuthSource:   "anthropic_api_key",
			DefaultModel: "claude-sonnet-4-6",
		},
		"claude-subscription": {
			Label:        "Claude Subscription",
			Provider:     "anthropic_claude",
			APIFormat:    "anthropic",
			AuthSource:   "claude_subscription",
			DefaultModel: "claude-sonnet-4-6",
		},
		"openai-compatible": {
			Label:        "OpenAI-Compatible API",
			Provider:     "openai",
			APIFormat:    "openai",
			AuthSource:   "openai_api_key",
			DefaultModel: "gpt-5.4",
		},
		"codex": {
			Label:        "Codex Subscription",
			Provider:     "openai_codex",
			APIFormat:    "openai",
			AuthSource:   "codex_subscription",
			DefaultModel: "gpt-5.4",
		},
		"copilot": {
			Label:        "GitHub Copilot",
			Provider:     "copilot",
			APIFormat:    "copilot",
			AuthSource:   "copilot_oauth",
			DefaultModel: "gpt-5.4",
		},
	}
}

// ResolvedModel returns the active model for this profile.
func (p ProviderProfile) ResolvedModel() string {
	if p.LastModel != "" {
		return p.LastModel
	}
	return p.DefaultModel
}

// ResolveProfile returns the active provider profile by name.
func (s Settings) ResolveProfile(name string) (string, ProviderProfile) {
	profiles := s.MergedProfiles()
	profileName := name
	if profileName == "" {
		profileName = s.ActiveProfile
	}
	if profileName == "" {
		profileName = "claude-api"
	}
	if p, ok := profiles[profileName]; ok {
		return profileName, p
	}
	// Fallback to claude-api default
	if p, ok := profiles["claude-api"]; ok {
		return "claude-api", p
	}
	return "claude-api", DefaultProviderProfiles()["claude-api"]
}

// MergedProfiles returns saved profiles merged over the built-in catalog.
func (s Settings) MergedProfiles() map[string]ProviderProfile {
	merged := DefaultProviderProfiles()
	for k, v := range s.Profiles {
		merged[k] = v
	}
	return merged
}

// ResolveAPIKey resolves the API key from settings, environment, or empty.
func (s Settings) ResolveAPIKey() string {
	if s.APIKey != "" {
		return s.APIKey
	}
	return ""
}

// MarshalJSON produces indented JSON output.
func (s Settings) MarshalJSON() ([]byte, error) {
	type Alias Settings
	return json.MarshalIndent((*Alias)(&s), "", "  ")
}
