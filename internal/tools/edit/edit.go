// Package edit implements the edit_file tool for performing string replacements in files.
package edit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

type editInput struct {
	Path       string `json:"path"`
	OldStr     string `json:"old_str"`
	NewStr     string `json:"new_str"`
	ReplaceAll bool   `json:"replace_all"`
}

// EditTool performs exact string replacements in files.
type EditTool struct{}

func (EditTool) Name() string { return "edit_file" }

func (EditTool) Description() string {
	return "Performs exact string replacements in files. Finds old_str and replaces it with new_str. " +
		"Use replace_all to replace all occurrences; otherwise the match must be unique."
}

func (EditTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path of the file to edit",
			},
			"old_str": map[string]any{
				"type":        "string",
				"description": "Text to find in the file",
			},
			"new_str": map[string]any{
				"type":        "string",
				"description": "Text to replace old_str with",
			},
			"replace_all": map[string]any{
				"type":        "boolean",
				"description": "Replace all occurrences of old_str",
				"default":     false,
			},
		},
		"required":             []string{"path", "old_str", "new_str"},
		"additionalProperties": false,
	}
}

func (EditTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input editInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	// Validate old_str is not empty
	if input.OldStr == "" {
		return tools.Result{Content: "old_str must not be empty", IsError: true}, nil
	}

	// Resolve path
	path := resolvePath(input.Path)

	// Check context
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return tools.Result{Content: fmt.Sprintf("File not found: %s", input.Path), IsError: true}, nil
		}
		return tools.Result{Content: fmt.Sprintf("Cannot read %s: %v", input.Path, err), IsError: true}, nil
	}

	content := string(data)

	// Count occurrences
	count := strings.Count(content, input.OldStr)
	if count == 0 {
		return tools.Result{Content: "old_str was not found in the file", IsError: true}, nil
	}

	// Check for ambiguous match when not replacing all
	if !input.ReplaceAll && count > 1 {
		return tools.Result{
			Content: fmt.Sprintf("old_str appears %d times in the file; use replace_all to replace all occurrences", count),
			IsError: true,
		}, nil
	}

	// Perform replacement
	var newContent string
	if input.ReplaceAll {
		newContent = strings.ReplaceAll(content, input.OldStr, input.NewStr)
	} else {
		newContent = strings.Replace(content, input.OldStr, input.NewStr, 1)
	}

	// Atomic write: write to temp file in same directory, then rename
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".edit_tmp_*")
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("Cannot create temp file: %v", err), IsError: true}, nil
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.WriteString(newContent); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return tools.Result{Content: fmt.Sprintf("Cannot write temp file: %v", err), IsError: true}, nil
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return tools.Result{Content: fmt.Sprintf("Cannot close temp file: %v", err), IsError: true}, nil
	}

	// Preserve original file permissions
	info, err := os.Stat(path)
	if err != nil {
		os.Remove(tmpPath)
		return tools.Result{Content: fmt.Sprintf("Cannot stat original file: %v", err), IsError: true}, nil
	}

	if err := os.Chmod(tmpPath, info.Mode()); err != nil {
		os.Remove(tmpPath)
		return tools.Result{Content: fmt.Sprintf("Cannot set file permissions: %v", err), IsError: true}, nil
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return tools.Result{Content: fmt.Sprintf("Cannot rename temp file: %v", err), IsError: true}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Updated %s", input.Path)}, nil
}

// resolvePath expands ~ to the home directory, resolves relative paths to absolute,
// and cleans the result.
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
