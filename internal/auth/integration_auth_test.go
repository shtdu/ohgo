//go:build integration

package auth_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/auth"
)

// EARS: REQ-AU-001
func TestIntegration_Auth_StoreAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := auth.NewManager(path)

	cred := &auth.Credential{
		Provider:  "anthropic",
		Kind:      "api_key",
		Value:     "sk-ant-test-key-123",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	err := mgr.Store(context.Background(), cred)
	require.NoError(t, err)

	loaded, err := mgr.Load(context.Background(), "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "anthropic", loaded.Provider)
	assert.Equal(t, "sk-ant-test-key-123", loaded.Value)
	assert.Equal(t, "api_key", loaded.Kind)
}

// EARS: REQ-AU-003
func TestIntegration_Auth_MultiProvider(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := auth.NewManager(path)

	// Store for multiple providers
	err := mgr.Store(context.Background(), &auth.Credential{
		Provider: "anthropic", Kind: "api_key", Value: "key-1",
		CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(),
	})
	require.NoError(t, err)

	err = mgr.Store(context.Background(), &auth.Credential{
		Provider: "openai", Kind: "api_key", Value: "key-2",
		CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(),
	})
	require.NoError(t, err)

	// Load each
	anthropic, err := mgr.Load(context.Background(), "anthropic")
	require.NoError(t, err)
	assert.Equal(t, "key-1", anthropic.Value)

	openai, err := mgr.Load(context.Background(), "openai")
	require.NoError(t, err)
	assert.Equal(t, "key-2", openai.Value)
}

// EARS: REQ-AU-004
func TestIntegration_Auth_StatusReporting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := auth.NewManager(path)

	err := mgr.Store(context.Background(), &auth.Credential{
		Provider: "anthropic", Kind: "api_key", Value: "sk-secret-key",
		CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(),
	})
	require.NoError(t, err)

	creds, err := mgr.List(context.Background())
	require.NoError(t, err)
	require.Len(t, creds, 1)
	assert.Equal(t, "anthropic", creds[0].Provider)
	// Value should be accessible (masking is done at display time)
}

// EARS: REQ-AU-001
func TestIntegration_Auth_Delete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := auth.NewManager(path)

	err := mgr.Store(context.Background(), &auth.Credential{
		Provider: "test", Kind: "api_key", Value: "to-delete",
		CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(),
	})
	require.NoError(t, err)

	err = mgr.Delete(context.Background(), "test")
	require.NoError(t, err)

	_, err = mgr.Load(context.Background(), "test")
	assert.Error(t, err, "loading deleted credential should fail")
}

// EARS: REQ-AU-001
func TestIntegration_Auth_ResolveKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	mgr := auth.NewManager(path)

	err := mgr.Store(context.Background(), &auth.Credential{
		Provider: "anthropic", Kind: "api_key", Value: "resolved-key",
		CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix(),
	})
	require.NoError(t, err)

	// ResolveKey maps auth sources to providers: "anthropic_api_key" -> "anthropic"
	key, err := mgr.ResolveKey(context.Background(), "anthropic_api_key")
	require.NoError(t, err)
	assert.Equal(t, "resolved-key", key)
}
