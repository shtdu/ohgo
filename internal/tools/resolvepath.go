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
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err == nil {
			path = abs
		}
	}
	return filepath.Clean(path)
}
