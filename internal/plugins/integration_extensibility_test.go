//go:build integration

package plugins_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/plugins"
)

// EARS: REQ-EX-001
func TestIntegration_Plugin_DiscoveryFromTempDir(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "test-plugin")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))

	manifest := `{
		"name": "test-plugin",
		"version": "1.0.0",
		"description": "A test plugin",
		"contributions": {}
	}`
	err := os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0o644)
	require.NoError(t, err)

	results, err := plugins.Discover(context.Background(), []string{dir})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "test-plugin", results[0].Manifest.Name)
	assert.Equal(t, "1.0.0", results[0].Manifest.Version)
}

// EARS: REQ-EX-001
func TestIntegration_Plugin_InvalidManifestSkipped(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "bad-plugin")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))

	err := os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(`{invalid json`), 0o644)
	require.NoError(t, err)

	results, err := plugins.Discover(context.Background(), []string{dir})
	require.NoError(t, err)
	assert.Empty(t, results, "invalid manifest should be skipped")
}

// EARS: REQ-EX-001
func TestIntegration_Plugin_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	results, err := plugins.Discover(context.Background(), []string{dir})
	require.NoError(t, err)
	assert.Empty(t, results)
}

// EARS: REQ-EX-001
func TestIntegration_Plugin_NestedPluginDir(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "my-plugin")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	manifest := `{"name": "nested-plugin", "version": "0.1.0"}`
	err := os.WriteFile(filepath.Join(subDir, "plugin.json"), []byte(manifest), 0o644)
	require.NoError(t, err)

	results, err := plugins.Discover(context.Background(), []string{dir})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "nested-plugin", results[0].Manifest.Name)
}

// EARS: REQ-EX-008
func TestIntegration_Plugin_Manager_EnableDisable(t *testing.T) {
	mgr := plugins.NewManager()
	assert.Empty(t, mgr.List())

	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "toggle-test")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))
	manifest := `{"name": "toggle-test", "version": "1.0.0"}`
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0o644))

	err := mgr.Discover(context.Background(), dir)
	require.NoError(t, err)

	list := mgr.List()
	assert.Len(t, list, 1)
	assert.Equal(t, "toggle-test", list[0].Manifest.Name)
}

// EARS: REQ-EX-008
func TestIntegration_Plugin_Manager_GetByName(t *testing.T) {
	mgr := plugins.NewManager()

	dir := t.TempDir()
	pluginDir := filepath.Join(dir, "get-test")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))
	manifest := `{"name": "get-test", "version": "2.0.0"}`
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(manifest), 0o644))

	err := mgr.Discover(context.Background(), dir)
	require.NoError(t, err)

	p := mgr.Get("get-test")
	require.NotNil(t, p)
	assert.Equal(t, "2.0.0", p.Manifest.Version)

	assert.Nil(t, mgr.Get("nonexistent"))
}
