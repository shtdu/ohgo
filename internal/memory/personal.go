package memory

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shtdu/ohgo/internal/config"
)

// PersonalDir returns the personal (user-level) memory directory.
// This is a fixed location (~/.ohgo/data/memory/_personal/) independent of
// the current working directory. It is created on demand.
func PersonalDir() (string, error) {
	dataDir, err := config.DataDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(dataDir, "memory", "_personal")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create personal memory dir: %w", err)
	}
	return dir, nil
}

// PersonalEntrypoint returns the path to the personal MEMORY.md index file.
func PersonalEntrypoint() (string, error) {
	dir, err := PersonalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "MEMORY.md"), nil
}
