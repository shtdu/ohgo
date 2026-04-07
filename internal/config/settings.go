package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Save persists settings to a JSON file.
func Save(s Settings, path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	return nil
}

func loadFromFile(path string) (Settings, error) {
	var s Settings
	data, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("parse %s: %w", path, err)
	}
	return s, nil
}

// mergeSettings applies override values on top of base settings.
func mergeSettings(base, override Settings) Settings {
	result := base

	if override.APIKey != "" {
		result.APIKey = override.APIKey
	}
	if override.Model != "" {
		result.Model = override.Model
	}
	if override.MaxTokens != 0 {
		result.MaxTokens = override.MaxTokens
	}
	if override.BaseURL != "" {
		result.BaseURL = override.BaseURL
	}
	if override.APIFormat != "" {
		result.APIFormat = override.APIFormat
	}
	if override.Provider != "" {
		result.Provider = override.Provider
	}
	if override.ActiveProfile != "" {
		result.ActiveProfile = override.ActiveProfile
	}
	if override.MaxTurns != 0 {
		result.MaxTurns = override.MaxTurns
	}
	if override.SystemPrompt != "" {
		result.SystemPrompt = override.SystemPrompt
	}
	if override.Theme != "" {
		result.Theme = override.Theme
	}
	if override.OutputStyle != "" {
		result.OutputStyle = override.OutputStyle
	}
	if override.VimMode {
		result.VimMode = true
	}
	if override.Verbose {
		result.Verbose = true
	}

	if override.Profiles != nil {
		if result.Profiles == nil {
			result.Profiles = make(map[string]ProviderProfile)
		}
		for k, v := range override.Profiles {
			result.Profiles[k] = v
		}
	}

	if override.Permission.Mode != "" {
		result.Permission.Mode = override.Permission.Mode
	}
	if len(override.Permission.AllowedTools) > 0 {
		result.Permission.AllowedTools = override.Permission.AllowedTools
	}
	if len(override.Permission.DeniedTools) > 0 {
		result.Permission.DeniedTools = override.Permission.DeniedTools
	}

	return result
}

func applyEnvOverrides(s *Settings) {
	if v := os.Getenv("ANTHROPIC_MODEL"); v != "" {
		s.Model = v
	} else if v := os.Getenv("OPENHARNESS_MODEL"); v != "" {
		s.Model = v
	}

	if v := os.Getenv("ANTHROPIC_BASE_URL"); v != "" {
		s.BaseURL = v
	} else if v := os.Getenv("OPENHARNESS_BASE_URL"); v != "" {
		s.BaseURL = v
	}

	if v := os.Getenv("OPENHARNESS_MAX_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.MaxTokens = n
		}
	}

	if v := os.Getenv("OPENHARNESS_MAX_TURNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.MaxTurns = n
		}
	}

	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		s.APIKey = v
	} else if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		s.APIKey = v
	}

	if v := os.Getenv("OPENHARNESS_API_FORMAT"); v != "" {
		s.APIFormat = v
	}

	if v := os.Getenv("OPENHARNESS_PROVIDER"); v != "" {
		s.Provider = v
	}
}

func parseBoolEnv(value string) bool {
	v := strings.TrimSpace(strings.ToLower(value))
	return v == "true" || v == "1" || v == "yes" || v == "on"
}
