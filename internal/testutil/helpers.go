//go:build integration

package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a temp directory cleaned up after the test.
func TempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

// WriteFile creates a file with the given content in the directory.
// Intermediate directories are created as needed.
func WriteFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// ReadFile reads a file relative to the directory.
func ReadFile(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

// MustJSON marshals v to JSON, failing the test on error.
func MustJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}
	return string(data)
}

// MustJSONIndent marshals v to pretty-printed JSON.
func MustJSONIndent(t *testing.T, v any) string {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}
	return string(data)
}
