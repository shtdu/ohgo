package auth

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_StoreAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := NewManager(path)

	cred := &Credential{
		Provider:  "anthropic",
		Kind:      "api_key",
		Value:     "sk-test-123",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	err := mgr.Store(context.Background(), cred)
	require.NoError(t, err)

	loaded, err := mgr.Load(context.Background(), "anthropic")
	require.NoError(t, assertCredential(cred, loaded))
}

func TestManager_Delete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := NewManager(path)

	cred := &Credential{
		Provider:  "openai",
		Kind:      "api_key",
		Value:     "sk-openai",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	require.NoError(t, mgr.Store(context.Background(), cred))
	require.NoError(t, mgr.Delete(context.Background(), "openai"))

	_, err := mgr.Load(context.Background(), "openai")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no credential")
}

func TestManager_List(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := NewManager(path)

	creds := []*Credential{
		{Provider: "anthropic", Kind: "api_key", Value: "sk-1", CreatedAt: 1, UpdatedAt: 1},
		{Provider: "openai", Kind: "api_key", Value: "sk-2", CreatedAt: 2, UpdatedAt: 2},
	}
	for _, c := range creds {
		require.NoError(t, mgr.Store(context.Background(), c))
	}

	list, err := mgr.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestManager_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := NewManager(path)

	cred := &Credential{Provider: "test", Kind: "api_key", Value: "key", CreatedAt: 1, UpdatedAt: 1}
	require.NoError(t, mgr.Store(context.Background(), cred))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestManager_ResolveKeyFromStore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := NewManager(path)

	cred := &Credential{Provider: "anthropic", Kind: "api_key", Value: "sk-stored", CreatedAt: 1, UpdatedAt: 1}
	require.NoError(t, mgr.Store(context.Background(), cred))

	key, err := mgr.ResolveKey(context.Background(), "anthropic_api_key")
	require.NoError(t, err)
	assert.Equal(t, "sk-stored", key)
}

func TestManager_ResolveKeyFromEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := NewManager(path)

	t.Setenv("OPENAI_API_KEY", "sk-env-key")

	key, err := mgr.ResolveKey(context.Background(), "openai_api_key")
	require.NoError(t, err)
	assert.Equal(t, "sk-env-key", key)
}

func TestManager_ResolveKeyStoreOverEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := NewManager(path)

	cred := &Credential{Provider: "openai", Kind: "api_key", Value: "sk-stored", CreatedAt: 1, UpdatedAt: 1}
	require.NoError(t, mgr.Store(context.Background(), cred))
	t.Setenv("OPENAI_API_KEY", "sk-env")

	// Stored credential should take priority.
	key, err := mgr.ResolveKey(context.Background(), "openai_api_key")
	require.NoError(t, err)
	assert.Equal(t, "sk-stored", key)
}

func TestManager_ResolveKeyMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := NewManager(path)

	_, err := mgr.ResolveKey(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no credential")
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(32)
	require.NoError(t, err)
	assert.Len(t, token, 64) // 32 bytes = 64 hex chars

	// Should be unique.
	token2, err := GenerateToken(32)
	require.NoError(t, err)
	assert.NotEqual(t, token, token2)
}

func TestAuthSourceToProvider(t *testing.T) {
	tests := []struct {
		authSource string
		want       string
	}{
		{"anthropic_api_key", "anthropic"},
		{"openai_api_key", "openai"},
		{"copilot_oauth", "copilot"},
		{"codex_subscription", "codex"},
		{"claude_subscription", "claude"},
		{"unknown", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, authSourceToProvider(tt.authSource))
	}
}

func assertCredential(want, got *Credential) error {
	if want.Provider != got.Provider || want.Kind != got.Kind || want.Value != got.Value {
		return fmt.Errorf("credential mismatch: want %+v, got %+v", want, got)
	}
	return nil
}
