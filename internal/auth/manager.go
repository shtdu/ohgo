package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Credential represents a stored authentication credential.
type Credential struct {
	Provider  string `json:"provider"`
	Kind      string `json:"kind"` // "api_key", "oauth_token", "external"
	Value     string `json:"value"`
	ExpiresAt int64  `json:"expires_at,omitempty"` // unix timestamp, 0 = never
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// Keyring abstracts secure credential storage.
type Keyring interface {
	Get(provider string) (*Credential, error)
	Set(cred *Credential) error
	Delete(provider string) error
	List() ([]*Credential, error)
}

// Manager handles credential storage and retrieval.
type Manager struct {
	storePath string
	keyring   Keyring
	mu        sync.RWMutex
}

// NewManager creates a new auth manager.
// If storePath is empty, defaults to ~/.openharness/credentials.json.
func NewManager(storePath string) *Manager {
	if storePath == "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			storePath = filepath.Join(home, ".openharness", "credentials.json")
		}
	}
	m := &Manager{
		storePath: storePath,
	}
	if storePath != "" {
		m.keyring = &fileKeyring{path: storePath}
	}
	return m
}

// Store saves a credential for the given provider.
func (m *Manager) Store(_ context.Context, cred *Credential) error {
	if m.keyring == nil {
		return fmt.Errorf("auth: no credential store configured")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.keyring.Set(cred)
}

// Load retrieves a credential for the given provider.
func (m *Manager) Load(_ context.Context, provider string) (*Credential, error) {
	if m.keyring == nil {
		return nil, fmt.Errorf("auth: no credential store configured")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.keyring.Get(provider)
}

// Delete removes a stored credential.
func (m *Manager) Delete(_ context.Context, provider string) error {
	if m.keyring == nil {
		return fmt.Errorf("auth: no credential store configured")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.keyring.Delete(provider)
}

// List returns all stored credentials (with values masked).
func (m *Manager) List(_ context.Context) ([]*Credential, error) {
	if m.keyring == nil {
		return nil, nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.keyring.List()
}

// ResolveKey returns an API key or token for the given auth source.
// Checks stored credentials first, then environment variables.
func (m *Manager) ResolveKey(ctx context.Context, authSource string) (string, error) {
	// Map auth sources to provider names for credential lookup.
	provider := authSourceToProvider(authSource)

	// Check stored credentials first.
	if provider != "" && m.keyring != nil {
		m.mu.RLock()
		cred, err := m.keyring.Get(provider)
		m.mu.RUnlock()
		if err == nil && cred != nil {
			return cred.Value, nil
		}
	}

	// Fall back to environment variable.
	if envKey := authSourceToEnv(authSource); envKey != "" {
		if val := os.Getenv(envKey); val != "" {
			return val, nil
		}
	}

	return "", fmt.Errorf("auth: no credential found for %q", authSource)
}

func authSourceToProvider(authSource string) string {
	switch authSource {
	case "anthropic_api_key":
		return "anthropic"
	case "openai_api_key":
		return "openai"
	case "copilot_oauth":
		return "copilot"
	case "codex_subscription":
		return "codex"
	case "claude_subscription":
		return "claude"
	default:
		return ""
	}
}

func authSourceToEnv(authSource string) string {
	switch authSource {
	case "anthropic_api_key":
		return "ANTHROPIC_API_KEY"
	case "openai_api_key":
		return "OPENAI_API_KEY"
	case "copilot_oauth":
		return "GITHUB_TOKEN"
	case "codex_subscription":
		return "CODEX_API_KEY"
	default:
		return ""
	}
}

// GenerateToken creates a cryptographically random hex token.
func GenerateToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

// --- File-based keyring implementation ---

type fileKeyring struct {
	path string
}

func (f *fileKeyring) load() (map[string]*Credential, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*Credential), nil
		}
		return nil, err
	}
	var store map[string]*Credential
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("auth: parse credentials: %w", err)
	}
	return store, nil
}

func (f *fileKeyring) save(store map[string]*Credential) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("auth: marshal credentials: %w", err)
	}
	// Ensure directory exists.
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("auth: create credentials dir: %w", err)
	}
	if err := os.WriteFile(f.path, data, 0o600); err != nil {
		return fmt.Errorf("auth: write credentials: %w", err)
	}
	return nil
}

func (f *fileKeyring) Get(provider string) (*Credential, error) {
	store, err := f.load()
	if err != nil {
		return nil, err
	}
	cred, ok := store[provider]
	if !ok {
		return nil, fmt.Errorf("auth: no credential for %q", provider)
	}
	return cred, nil
}

func (f *fileKeyring) Set(cred *Credential) error {
	store, err := f.load()
	if err != nil {
		return err
	}
	store[cred.Provider] = cred
	return f.save(store)
}

func (f *fileKeyring) Delete(provider string) error {
	store, err := f.load()
	if err != nil {
		return err
	}
	if _, ok := store[provider]; !ok {
		return fmt.Errorf("auth: no credential for %q", provider)
	}
	delete(store, provider)
	return f.save(store)
}

func (f *fileKeyring) List() ([]*Credential, error) {
	store, err := f.load()
	if err != nil {
		return nil, err
	}
	var result []*Credential
	for _, cred := range store {
		result = append(result, cred)
	}
	return result, nil
}
