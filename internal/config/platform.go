package config

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// PlatformName identifies the runtime platform.
type PlatformName string

const (
	PlatformMacOS   PlatformName = "macos"
	PlatformLinux   PlatformName = "linux"
	PlatformWindows PlatformName = "windows"
	PlatformWSL     PlatformName = "wsl"
	PlatformUnknown PlatformName = "unknown"
)

// PlatformCapabilities describes what the current platform supports.
type PlatformCapabilities struct {
	Name                     PlatformName
	SupportsPosixShell       bool
	SupportsNativeWinShell   bool
	SupportsTmux             bool
	SupportsSwarmMailbox     bool
	SupportsSandboxRuntime   bool
}

// DetectPlatform returns the normalized platform name.
func DetectPlatform() PlatformName {
	switch runtime.GOOS {
	case "darwin":
		return PlatformMacOS
	case "windows":
		return PlatformWindows
	case "linux":
		if isWSL() {
			return PlatformWSL
		}
		return PlatformLinux
	default:
		return PlatformUnknown
	}
}

// GetPlatformCapabilities returns the capability matrix for a platform.
func GetPlatformCapabilities(name PlatformName) PlatformCapabilities {
	switch name {
	case PlatformMacOS, PlatformLinux, PlatformWSL:
		return PlatformCapabilities{
			Name:                   name,
			SupportsPosixShell:     true,
			SupportsNativeWinShell: false,
			SupportsTmux:           true,
			SupportsSwarmMailbox:   true,
			SupportsSandboxRuntime: true,
		}
	case PlatformWindows:
		return PlatformCapabilities{
			Name:                   name,
			SupportsPosixShell:     false,
			SupportsNativeWinShell: true,
			SupportsTmux:           false,
			SupportsSwarmMailbox:   false,
			SupportsSandboxRuntime: false,
		}
	default:
		return PlatformCapabilities{
			Name: name,
		}
	}
}

// Shell returns the user's preferred shell.
func Shell() string {
	if sh := os.Getenv("SHELL"); sh != "" {
		return sh
	}
	if comspec := os.Getenv("COMSPEC"); comspec != "" {
		return comspec
	}
	return "/bin/sh"
}

// WorkingDir returns the current working directory.
func WorkingDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get working directory: %w", err)
	}
	return dir, nil
}

// isWSL detects Windows Subsystem for Linux.
func isWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != "" {
		return true
	}
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

// HasCommand checks if a command is available on PATH.
func HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
