// Package write implements the write_file tool for writing file contents.
package write

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

type writeInput struct {
	Path             string `json:"path"`
	Content          string `json:"content"`
	CreateDirectories bool   `json:"create_directories"`
}

// WriteTool writes content to a file, creating it or overwriting if it exists.
type WriteTool struct{}

func (WriteTool) Name() string { return "write_file" }

func (WriteTool) Description() string {
	return "Writes content to a file on the local filesystem. Creates the file if it does not exist, or overwrites it if it does."
}

func (WriteTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path of the file to write",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Full contents to write to the file",
			},
			"create_directories": map[string]any{
				"type":        "boolean",
				"description": "Automatically create parent directories if they do not exist",
				"default":     true,
			},
		},
		"required":             []string{"path", "content"},
		"additionalProperties": false,
	}
}

func (WriteTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input writeInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	// Resolve path
	path := resolvePath(input.Path)

	// Check context
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	// Check if path is an existing directory
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return tools.Result{Content: fmt.Sprintf("Cannot write to directory: %s", input.Path), IsError: true}, nil
	}

	// Create parent directories if needed
	createDirs := input.CreateDirectories // zero value is false
	if !createDirs {
		// Default to true when the field is not explicitly set.
		// We detect this by re-unmarshalling into a raw map to check
		// whether the key was present.
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(args, &raw)
		if _, ok := raw["create_directories"]; !ok {
			createDirs = true
		}
	}

	if createDirs {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return tools.Result{Content: fmt.Sprintf("Cannot create directories: %v", err), IsError: true}, nil
		}
	}

	// Write file
	if err := os.WriteFile(path, []byte(input.Content), 0644); err != nil {
		return tools.Result{Content: fmt.Sprintf("Cannot write file: %v", err), IsError: true}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Wrote %s", path)}, nil
}

func resolvePath(path string) string {
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
