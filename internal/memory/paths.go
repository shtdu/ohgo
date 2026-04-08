package memory

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shtdu/ohgo/internal/config"
)

// ProjectDir returns the persistent memory directory for a project,
// derived from the SHA1 hash of the absolute working directory path.
func ProjectDir(cwd string) (string, error) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		return "", fmt.Errorf("resolve cwd: %w", err)
	}
	digest := sha1.Sum([]byte(abs))
	base := filepath.Base(abs)
	dirName := fmt.Sprintf("%s-%x", base, digest[:6])

	dataDir, err := config.DataDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(dataDir, "memory", dirName)
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
