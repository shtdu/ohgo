package config

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectPlatform(t *testing.T) {
	p := DetectPlatform()
	switch runtime.GOOS {
	case "darwin":
		assert.Equal(t, PlatformMacOS, p)
	case "linux":
		// could be linux or wsl depending on environment
		assert.Contains(t, []PlatformName{PlatformLinux, PlatformWSL}, p)
	case "windows":
		assert.Equal(t, PlatformWindows, p)
	}
}

func TestGetPlatformCapabilities_MacOS(t *testing.T) {
	caps := GetPlatformCapabilities(PlatformMacOS)
	assert.Equal(t, PlatformMacOS, caps.Name)
	assert.True(t, caps.SupportsPosixShell)
	assert.False(t, caps.SupportsNativeWinShell)
	assert.True(t, caps.SupportsTmux)
}

func TestGetPlatformCapabilities_Windows(t *testing.T) {
	caps := GetPlatformCapabilities(PlatformWindows)
	assert.Equal(t, PlatformWindows, caps.Name)
	assert.False(t, caps.SupportsPosixShell)
	assert.True(t, caps.SupportsNativeWinShell)
	assert.False(t, caps.SupportsTmux)
}

func TestGetPlatformCapabilities_Unknown(t *testing.T) {
	caps := GetPlatformCapabilities(PlatformUnknown)
	assert.Equal(t, PlatformUnknown, caps.Name)
	assert.False(t, caps.SupportsPosixShell)
}

func TestShell(t *testing.T) {
	sh := Shell()
	assert.NotEmpty(t, sh)
}

func TestWorkingDir(t *testing.T) {
	dir, err := WorkingDir()
	require.NoError(t, err)
	assert.NotEmpty(t, dir)
}

func TestIsWSL_FalseOnMac(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("only runs on macOS")
	}
	assert.False(t, isWSL())
}

func TestIsWSL_WithEnvVar(t *testing.T) {
	t.Setenv("WSL_DISTRO_NAME", "Ubuntu")
	assert.True(t, isWSL())
}

func TestHasCommand(t *testing.T) {
	assert.True(t, HasCommand("go"))
	assert.False(t, HasCommand("nonexistent_command_12345"))
}
