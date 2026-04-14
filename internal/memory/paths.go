package memory

import (
	"fmt"
	"os"
	"path/filepath"
)

// ProjectDir returns the project-scoped memory directory at <cwd>/.ohgo/data/memory/.
// This places memory files alongside the project, making them portable with the repo.
func ProjectDir(cwd string) (string, error) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return "", fmt.Errorf("resolve cwd: %w", err)
	}
	dir := filepath.Join(abs, ".ohgo", "data", "memory")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create memory dir: %w", err)
	}
	return dir, nil
}

// Entrypoint returns the project memory index file path (MEMORY.md).
func Entrypoint(cwd string) (string, error) {
	dir, err := ProjectDir(cwd)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "MEMORY.md"), nil
}
