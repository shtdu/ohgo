package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallCopiesFiles(t *testing.T) {
	// Set up a fake config dir via environment variable.
	configDir := t.TempDir()
	t.Setenv("OHGO_CONFIG_DIR", configDir)

	// Create a source plugin directory.
	src := t.TempDir()
	srcPlugin := filepath.Join(src, "my-plugin")
	require.NoError(t, os.MkdirAll(filepath.Join(srcPlugin, "skills"), 0o755))

	manifest := `{"name": "my-plugin", "version": "1.0.0"}`
	require.NoError(t, os.WriteFile(filepath.Join(srcPlugin, "plugin.json"), []byte(manifest), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(srcPlugin, "skills", "hello.md"), []byte("# Hello"), 0o644))

	dest, err := Install(srcPlugin)
	require.NoError(t, err)

	expectedDest := filepath.Join(configDir, "plugins", "my-plugin")
	assert.Equal(t, expectedDest, dest)

	// Verify files were copied.
	data, err := os.ReadFile(filepath.Join(dest, "plugin.json"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "my-plugin")

	data, err = os.ReadFile(filepath.Join(dest, "skills", "hello.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Hello", string(data))
}

func TestInstallOverwritesExisting(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("OHGO_CONFIG_DIR", configDir)

	src := t.TempDir()
	srcPlugin := filepath.Join(src, "overwrite-test")
	require.NoError(t, os.MkdirAll(srcPlugin, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(srcPlugin, "plugin.json"), []byte(`{"name":"overwrite-test","version":"2.0.0"}`), 0o644))

	// First install.
	_, err := Install(srcPlugin)
	require.NoError(t, err)

	// Modify source and reinstall.
	require.NoError(t, os.WriteFile(filepath.Join(srcPlugin, "plugin.json"), []byte(`{"name":"overwrite-test","version":"3.0.0"}`), 0o644))

	dest, err := Install(srcPlugin)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dest, "plugin.json"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "3.0.0")
}

func TestInstallFailsOnFileSource(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("OHGO_CONFIG_DIR", configDir)

	src := t.TempDir()
	filePath := filepath.Join(src, "not-a-dir.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("hello"), 0o644))

	_, err := Install(filePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestUninstallRemovesDirectory(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("OHGO_CONFIG_DIR", configDir)

	// Create a plugin to uninstall.
	pluginDir := filepath.Join(configDir, "plugins", "to-remove")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(`{}`), 0o644))

	removed, err := Uninstall("to-remove")
	require.NoError(t, err)
	assert.True(t, removed)

	// Verify directory is gone.
	_, statErr := os.Stat(pluginDir)
	assert.True(t, os.IsNotExist(statErr))
}

func TestUninstallMissingReturnsFalse(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("OHGO_CONFIG_DIR", configDir)

	removed, err := Uninstall("nonexistent")
	require.NoError(t, err)
	assert.False(t, removed)
}

func TestPluginsDirCreatesDirectory(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("OHGO_CONFIG_DIR", configDir)

	// Plugins dir should not exist yet.
	pluginsPath := filepath.Join(configDir, "plugins")
	_, err := os.Stat(pluginsPath)
	assert.True(t, os.IsNotExist(err))

	dir, err := PluginsDir()
	require.NoError(t, err)
	assert.Equal(t, pluginsPath, dir)

	// Now it should exist.
	info, err := os.Stat(pluginsPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
