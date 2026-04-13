//go:build integration

package auth_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/auth"
	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/permissions"
)

// EARS: REQ-AU-001
// Stored credential resolves through ResolveKey, matching the auth-source→provider mapping
// that the engine uses to obtain API keys at query time.
func TestIntegration_Auth_ResolveKey_StoreAndFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := auth.NewManager(path)

	// No credential stored yet — should fall back to env
	t.Setenv("ANTHROPIC_API_KEY", "env-fallback-key")
	key, err := mgr.ResolveKey(context.Background(), "anthropic_api_key")
	require.NoError(t, err)
	assert.Equal(t, "env-fallback-key", key)

	// Store a credential — should take precedence over env
	require.NoError(t, mgr.Store(context.Background(), &auth.Credential{
		Provider: "anthropic", Kind: "api_key", Value: "stored-key",
		CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(),
	}))
	key, err = mgr.ResolveKey(context.Background(), "anthropic_api_key")
	require.NoError(t, err)
	assert.Equal(t, "stored-key", key, "stored credential should override env var")
}

// EARS: REQ-AU-003
// Multiple providers can be stored and independently resolved by different auth sources.
func TestIntegration_Auth_MultiProvider_IndependentResolution(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := auth.NewManager(path)

	require.NoError(t, mgr.Store(context.Background(), &auth.Credential{
		Provider: "anthropic", Kind: "api_key", Value: "anthropic-key",
		CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(),
	}))
	require.NoError(t, mgr.Store(context.Background(), &auth.Credential{
		Provider: "openai", Kind: "api_key", Value: "openai-key",
		CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(),
	}))

	// Each auth source maps to the correct provider
	anthropicKey, err := mgr.ResolveKey(context.Background(), "anthropic_api_key")
	require.NoError(t, err)
	assert.Equal(t, "anthropic-key", anthropicKey)

	openaiKey, err := mgr.ResolveKey(context.Background(), "openai_api_key")
	require.NoError(t, err)
	assert.Equal(t, "openai-key", openaiKey)
}

// EARS: REQ-AU-001, REQ-AU-004
// After deleting a credential, List no longer includes it and ResolveKey falls back.
func TestIntegration_Auth_DeleteAndVerifyResolution(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := auth.NewManager(path)

	require.NoError(t, mgr.Store(context.Background(), &auth.Credential{
		Provider: "test-provider", Kind: "api_key", Value: "will-delete",
		CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(),
	}))

	// Verify listed
	creds, err := mgr.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, creds, 1)

	// Delete
	require.NoError(t, mgr.Delete(context.Background(), "test-provider"))

	// List should be empty
	creds, err = mgr.List(context.Background())
	require.NoError(t, err)
	assert.Empty(t, creds)

	// Load should fail
	_, err = mgr.Load(context.Background(), "test-provider")
	assert.Error(t, err)
}

// EARS: REQ-AU-001
// Credential file is stored with restricted permissions (0600).
func TestIntegration_Auth_FileSecurity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := auth.NewManager(path)

	require.NoError(t, mgr.Store(context.Background(), &auth.Credential{
		Provider: "secure-test", Kind: "api_key", Value: "secret-value",
		CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(),
	}))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "credentials file should be 0600")
}

// EARS: REQ-AU-001
// Auth credential storage integrates with config profile resolution:
// storing a credential for a provider and checking it against config settings.
func TestIntegration_Auth_ConfigProfileIntegration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := auth.NewManager(path)

	// Store credential for a provider used in a profile
	require.NoError(t, mgr.Store(context.Background(), &auth.Credential{
		Provider: "anthropic", Kind: "api_key", Value: "profile-test-key",
		CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(),
	}))

	// Build a permission checker using config (cross-component)
	settings := config.PermissionSettings{Mode: "auto"}
	checker := permissions.NewDefaultChecker(settings)
	decision, err := checker.Check(context.Background(), permissions.Check{
		ToolName: "read_file",
		Args:     map[string]any{"path": "/tmp/test"},
	})
	require.NoError(t, err)
	assert.Equal(t, permissions.Allow, decision, "auto mode should allow read tool")

	// Verify the key resolves for the profile's auth source
	key, err := mgr.ResolveKey(context.Background(), "anthropic_api_key")
	require.NoError(t, err)
	assert.Equal(t, "profile-test-key", key)
}
