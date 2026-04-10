package plugins

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/shtdu/ohgo/internal/config"
)

// PluginsDir returns the user plugins directory, creating it if needed.
func PluginsDir() (string, error) {
	cfgDir, err := config.ConfigDir()
	if err != nil {
		return "", fmt.Errorf("get config dir: %w", err)
	}
	dir := filepath.Join(cfgDir, "plugins")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create plugins dir %s: %w", dir, err)
	}
	return dir, nil
}

// Install copies a plugin directory into the user plugins directory.
// Returns the destination path.
func Install(source string) (string, error) {
	src, err := filepath.Abs(source)
	if err != nil {
		return "", fmt.Errorf("resolve source path: %w", err)
	}

	info, err := os.Stat(src)
	if err != nil {
		return "", fmt.Errorf("stat source %s: %w", src, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("source %s is not a directory", src)
	}

	destDir, err := PluginsDir()
	if err != nil {
		return "", err
	}

	dest := filepath.Join(destDir, filepath.Base(src))

	// Remove existing destination if it exists.
	if _, err := os.Stat(dest); err == nil {
		if err := os.RemoveAll(dest); err != nil {
			return "", fmt.Errorf("remove existing plugin %s: %w", dest, err)
		}
	}

	if err := copyDir(src, dest); err != nil {
		return "", fmt.Errorf("copy plugin: %w", err)
	}

	return dest, nil
}

// Uninstall removes a plugin directory by name from the user plugins directory.
// Returns true if the plugin was removed, false if it was not found.
func Uninstall(name string) (bool, error) {
	destDir, err := PluginsDir()
	if err != nil {
		return false, err
	}

	path := filepath.Join(destDir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}

	if err := os.RemoveAll(path); err != nil {
		return false, fmt.Errorf("remove plugin %s: %w", path, err)
	}

	return true, nil
}

// copyDir recursively copies a directory tree.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target, info.Mode())
	})
}

// copyFile copies a single file.
func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer func() { _ = in.Close() }()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy data: %w", err)
	}

	return nil
}
