package tools

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolvePath expands ~ to the home directory, resolves relative paths to
// absolute, and cleans the result.
func ResolvePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fall back to the HOME environment variable.
			home = os.Getenv("HOME")
		}
		if home != "" {
			path = filepath.Join(home, path[2:])
		}
	}
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err == nil {
			path = abs
		}
	}
	return filepath.Clean(path)
}
