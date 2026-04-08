// Package read implements the read_file tool for reading file contents.
package read

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

const (
	defaultLimit = 200
	maxLimit     = 2000
	binaryCheckSize = 8192
)

type readInput struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

// ReadTool reads file contents with line numbers.
type ReadTool struct{}

func (ReadTool) Name() string { return "read_file" }

func (ReadTool) Description() string {
	return "Reads a file from the local filesystem. Returns content with line numbers."
}

func (ReadTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path of the file to read",
			},
			"offset": map[string]any{
				"type":        "integer",
				"description": "Zero-based starting line number",
				"default":     0,
				"minimum":     0,
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Number of lines to return",
				"default":     defaultLimit,
				"minimum":     1,
				"maximum":     maxLimit,
			},
		},
		"required":             []string{"path"},
		"additionalProperties": false,
	}
}

func (ReadTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input readInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	// Resolve path
	path := expandPath(input.Path)

	// Check context
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	// Stat the path
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return tools.Result{Content: fmt.Sprintf("File not found: %s", input.Path), IsError: true}, nil
		}
		return tools.Result{Content: fmt.Sprintf("Cannot access %s: %v", input.Path, err), IsError: true}, nil
	}

	if info.IsDir() {
		return tools.Result{Content: fmt.Sprintf("Cannot read directory: %s", input.Path), IsError: true}, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("Cannot read %s: %v", input.Path, err), IsError: true}, nil
	}

	// Binary detection
	if isBinary(data) {
		return tools.Result{Content: fmt.Sprintf("Binary file cannot be read as text: %s", input.Path), IsError: true}, nil
	}

	// Split into lines
	content := string(data)
	lines := strings.Split(content, "\n")

	// Remove trailing empty line from final newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Empty file
	if len(lines) == 0 {
		return tools.Result{Content: ""}, nil
	}

	// Apply offset and limit
	offset := max(input.Offset, 0)
	if offset >= len(lines) {
		return tools.Result{Content: "(no content in selected range)"}, nil
	}

	limit := input.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	end := min(offset+limit, len(lines))

	// Build output with line numbers (1-based)
	var buf strings.Builder
	for i := offset; i < end; i++ {
		fmt.Fprintf(&buf, "%d\t%s\n", i+1, lines[i])
	}

	return tools.Result{Content: buf.String()}, nil
}

func isBinary(data []byte) bool {
	check := data
	if len(check) > binaryCheckSize {
		check = check[:binaryCheckSize]
	}
	return bytes.ContainsRune(check, 0)
}

func expandPath(path string) string {
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
