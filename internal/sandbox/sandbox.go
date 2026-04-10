// Package sandbox provides sandbox runtime integration for command execution.
package sandbox

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Availability describes whether sandbox execution is available.
type Availability struct {
	Enabled   bool
	Available bool
	Reason    string
	Command   string // path to srt binary, empty if not found
}

// Active returns true if sandbox can be used.
func (a Availability) Active() bool {
	return a.Enabled && a.Available
}

// CheckAvailability checks whether the sandbox runtime is installed and usable.
func CheckAvailability() Availability {
	// 1. Check if srt binary is on PATH
	srtPath, err := exec.LookPath("srt")
	if err != nil {
		return Availability{
			Enabled:   false,
			Available: false,
			Reason:    "srt binary not found on PATH",
		}
	}

	// 2. Platform-specific checks
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("bwrap"); err != nil {
			return Availability{
				Enabled:   false,
				Available: false,
				Reason:    "bwrap not found (required on Linux)",
				Command:   srtPath,
			}
		}
	case "darwin":
		// sandbox-exec availability is optional on macOS
	}

	return Availability{
		Enabled:   true,
		Available: true,
		Command:   srtPath,
	}
}

// WrapCommand wraps a command for sandboxed execution.
// Returns the wrapped argv, a temp config file path (caller must clean up), or error.
// If sandbox is not active, returns the original argv unchanged.
func WrapCommand(argv []string) ([]string, string, error) {
	avail := CheckAvailability()
	if !avail.Active() {
		return argv, "", nil
	}

	// Build config
	config, err := buildConfig()
	if err != nil {
		return nil, "", fmt.Errorf("build sandbox config: %w", err)
	}

	// Write config to temp file
	tmpFile, err := os.CreateTemp("", "sandbox-config-*.json")
	if err != nil {
		return nil, "", fmt.Errorf("create temp config: %w", err)
	}
	defer func() { _ = tmpFile.Close() }()

	if _, err := tmpFile.Write(config); err != nil {
		_ = os.Remove(tmpFile.Name())
		return nil, "", fmt.Errorf("write sandbox config: %w", err)
	}

	// Build wrapped command: srt --settings <config> -- <original command>
	wrapped := []string{
		avail.Command,
		"--settings", tmpFile.Name(),
		"--",
	}
	wrapped = append(wrapped, argv...)

	return wrapped, tmpFile.Name(), nil
}

// buildConfig creates a simplified sandbox runtime configuration.
func buildConfig() ([]byte, error) {
	cfg := map[string]any{
		"allow": []string{},
		"deny":  []string{},
	}
	return json.MarshalIndent(cfg, "", "  ")
}
