package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const defaultCodexBaseURL = "http://localhost:8967/v1/chat/completions"

// extractCodexToken reads the Codex CLI's local credentials.
// Returns (token, baseURL, error).
func extractCodexToken() (string, string, error) {
	// Check environment variables first.
	if token := os.Getenv("CODEX_TOKEN"); token != "" {
		baseURL := os.Getenv("CODEX_API_URL")
		if baseURL == "" {
			baseURL = defaultCodexBaseURL
		}
		return token, baseURL, nil
	}

	// Fall back to ~/.codex/credentials.json.
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("codex: find home dir: %w", err)
	}

	credPath := filepath.Join(home, ".codex", "credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		return "", "", fmt.Errorf("codex: read credentials: %w (install Codex CLI or set CODEX_TOKEN)", err)
	}

	var credData map[string]any
	if err := json.Unmarshal(data, &credData); err != nil {
		return "", "", fmt.Errorf("codex: parse credentials: %w", err)
	}

	token, _ := credData["token"].(string)
	if token == "" {
		token, _ = credData["api_key"].(string)
	}
	if token == "" {
		return "", "", fmt.Errorf("codex: no token found in %s", credPath)
	}

	baseURL := defaultCodexBaseURL
	if url, ok := credData["api_url"].(string); ok && url != "" {
		baseURL = url
	}

	return token, baseURL, nil
}
