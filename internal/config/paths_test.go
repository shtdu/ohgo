package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDir_Default(t *testing.T) {
	t.Setenv("OPENHARNESS_CONFIG_DIR", "")
	dir, err := ConfigDir()
	require.NoError(t, err)
	home, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(home, ".openharness"), dir)
}

func TestConfigDir_EnvOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	dir, err := ConfigDir()
	require.NoError(t, err)
	assert.Equal(t, tmp, dir)
}

func TestConfigFilePath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	path, err := ConfigFilePath()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, "settings.json"), path)
}

func TestDataDir_Default(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("OPENHARNESS_DATA_DIR", "")
	dir, err := DataDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, "data"), dir)
}

func TestDataDir_EnvOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_DATA_DIR", tmp)
	dir, err := DataDir()
	require.NoError(t, err)
	assert.Equal(t, tmp, dir)
}

func TestSessionsDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmp)
	t.Setenv("OPENHARNESS_DATA_DIR", "")
	dir, err := SessionsDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, "data", "sessions"), dir)
	_, err = os.Stat(dir)
	assert.NoError(t, err, "sessions dir should exist")
}

func TestProjectDir(t *testing.T) {
	tmp := t.TempDir()
	dir, err := ProjectDir(tmp)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmp, ".openharness"), dir)
	_, err = os.Stat(dir)
	assert.NoError(t, err, "project dir should exist")
}
