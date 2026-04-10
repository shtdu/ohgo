package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Flow represents an authentication method.
type Flow interface {
	// Name returns the flow identifier (e.g. "api_key", "device_code", "external_cli").
	Name() string
	// Authenticate executes the auth flow and returns a credential.
	Authenticate(ctx context.Context) (*Credential, error)
}

// APIKeyFlow prompts the user for an API key.
type APIKeyFlow struct {
	Provider string
}

// Name returns the flow name.
func (f *APIKeyFlow) Name() string { return "api_key" }

// Authenticate reads an API key from stdin and returns a credential.
func (f *APIKeyFlow) Authenticate(ctx context.Context) (*Credential, error) {
	fmt.Printf("Enter API key for %s: ", f.Provider)

	// Read from stdin.
	var key string
	ch := make(chan string, 1)
	go func() {
		_, _ = fmt.Fscanln(os.Stdin, &key)
		ch <- key
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case k := <-ch:
		if k == "" {
			return nil, fmt.Errorf("auth: empty API key")
		}
		now := time.Now().Unix()
		return &Credential{
			Provider:  f.Provider,
			Kind:      "api_key",
			Value:     k,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}
}

// ExternalCLIFlow reads credentials from an external CLI's config file.
type ExternalCLIFlow struct {
	Provider  string
	ConfigDir string // e.g. "~/.codex"
	FileName  string // e.g. "credentials.json"
}

// Name returns the flow name.
func (f *ExternalCLIFlow) Name() string { return "external_cli" }

// Authenticate reads the external CLI's credential file.
func (f *ExternalCLIFlow) Authenticate(_ context.Context) (*Credential, error) {
	dir := expandHome(f.ConfigDir)
	path := filepath.Join(dir, f.FileName)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("auth: read %s: %w", path, err)
	}

	var credData map[string]any
	if err := json.Unmarshal(data, &credData); err != nil {
		return nil, fmt.Errorf("auth: parse %s: %w", path, err)
	}

	token, _ := credData["token"].(string)
	if token == "" {
		token, _ = credData["api_key"].(string)
	}
	if token == "" {
		return nil, fmt.Errorf("auth: no token found in %s", path)
	}

	now := time.Now().Unix()
	return &Credential{
		Provider:  f.Provider,
		Kind:      "external",
		Value:     token,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// SelectFlow returns the appropriate auth flow for the given auth source.
func SelectFlow(authSource string) Flow {
	switch authSource {
	case "anthropic_api_key":
		return &APIKeyFlow{Provider: "anthropic"}
	case "openai_api_key":
		return &APIKeyFlow{Provider: "openai"}
	case "copilot_oauth":
		return NewCopilotDeviceFlow()
	case "codex_subscription":
		return &ExternalCLIFlow{
			Provider:  "codex",
			ConfigDir: "~/.codex",
			FileName:  "credentials.json",
		}
	case "claude_subscription":
		return &ExternalCLIFlow{
			Provider:  "claude",
			ConfigDir: "~/.claude",
			FileName:  ".credentials.json",
		}
	default:
		return nil
	}
}

func expandHome(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}
