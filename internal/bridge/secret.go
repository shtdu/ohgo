package bridge

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Secret represents a work secret used to authenticate bridge sessions.
type Secret struct {
	ID         string `json:"id"`
	Value      string `json:"value"`
	BridgeName string `json:"bridge_name"`
	CreatedAt  int64  `json:"created_at"`
	ExpiresAt  int64  `json:"expires_at,omitempty"`
}

// SecretManager manages work secrets for bridge sessions.
type SecretManager struct {
	storePath string
	mu        sync.Mutex
}

// NewSecretManager creates a new secret manager.
func NewSecretManager(storePath string) *SecretManager {
	return &SecretManager{storePath: storePath}
}

// Generate creates a new work secret for the given bridge.
func (sm *SecretManager) Generate(bridgeName string) (*Secret, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	id, err := generateHex(16)
	if err != nil {
		return nil, err
	}
	value, err := generateHex(32)
	if err != nil {
		return nil, err
	}

	secret := &Secret{
		ID:         id,
		Value:      value,
		BridgeName: bridgeName,
		CreatedAt:  time.Now().Unix(),
	}

	if err := sm.save(secret); err != nil {
		return nil, err
	}
	return secret, nil
}

// Validate checks if a secret is valid and not expired.
func (sm *SecretManager) Validate(secretID, secretValue string) (bool, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	store, err := sm.load()
	if err != nil {
		return false, err
	}

	secret, ok := store[secretID]
	if !ok {
		return false, nil
	}

	if secret.Value != secretValue {
		return false, nil
	}

	// Check expiry (0 = never expires).
	if secret.ExpiresAt > 0 && time.Now().Unix() > secret.ExpiresAt {
		return false, nil
	}

	return true, nil
}

// Rotate generates a new secret, invalidating the old one.
func (sm *SecretManager) Rotate(secretID string) (*Secret, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	store, err := sm.load()
	if err != nil {
		return nil, err
	}

	old, ok := store[secretID]
	if !ok {
		return nil, fmt.Errorf("secret %q not found", secretID)
	}

	newValue, err := generateHex(32)
	if err != nil {
		return nil, err
	}

	old.Value = newValue
	old.CreatedAt = time.Now().Unix()

	if err := sm.saveAll(store); err != nil {
		return nil, err
	}
	return old, nil
}

// Revoke removes a secret.
func (sm *SecretManager) Revoke(secretID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	store, err := sm.load()
	if err != nil {
		return err
	}

	if _, ok := store[secretID]; !ok {
		return fmt.Errorf("secret %q not found", secretID)
	}

	delete(store, secretID)
	return sm.saveAll(store)
}

func (sm *SecretManager) load() (map[string]*Secret, error) {
	data, err := os.ReadFile(sm.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*Secret), nil
		}
		return nil, err
	}
	var store map[string]*Secret
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return store, nil
}

func (sm *SecretManager) saveAll(store map[string]*Secret) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(sm.storePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(sm.storePath, data, 0o600)
}

func (sm *SecretManager) save(secret *Secret) error {
	store, err := sm.load()
	if err != nil {
		return err
	}
	store[secret.ID] = secret
	return sm.saveAll(store)
}

func generateHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
