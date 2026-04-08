// Package glob implements the glob tool for file pattern matching.
package glob

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

const (
	defaultGlobLimit = 200
	maxGlobLimit     = 5000
)

type globInput struct {
	Pattern string `json:"pattern"`
	Limit   int    `json:"limit"`
}

// GlobTool finds files matching a glob pattern.
type GlobTool struct{}

func (GlobTool) Name() string { return "glob" }

func (GlobTool) Description() string {
	return "Fast file pattern matching tool. Supports patterns like '**/*.go' to find files recursively."
}

func (GlobTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Glob pattern (e.g. '**/*.go', 'src/*.ts')",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results",
				"default":     defaultGlobLimit,
				"minimum":     1,
				"maximum":     maxGlobLimit,
			},
		},
		"required":             []string{"pattern"},
		"additionalProperties": false,
	}
}

func (GlobTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input globInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.Pattern == "" {
		return tools.Result{Content: "pattern is required", IsError: true}, nil
	}

	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	limit := input.Limit
	if limit <= 0 {
		limit = defaultGlobLimit
	}
	limit = min(limit, maxGlobLimit)

	cwd, err := os.Getwd()
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("cannot get working directory: %v", err), IsError: true}, nil
	}

	var matches []string

	if strings.Contains(input.Pattern, "**") {
		matches = globRecursive(cwd, input.Pattern, limit)
	} else {
		// Use filepath.Glob for simple patterns
		raw, err := filepath.Glob(input.Pattern)
		if err != nil {
			return tools.Result{Content: fmt.Sprintf("invalid pattern: %v", err), IsError: true}, nil
		}
		for _, m := range raw {
			rel, err := filepath.Rel(cwd, m)
			if err != nil {
				rel = m
			}
			matches = append(matches, rel)
			if len(matches) >= limit {
				break
			}
		}
	}

	if len(matches) == 0 {
		return tools.Result{Content: "(no matches)"}, nil
	}

	sort.Strings(matches)
	return tools.Result{Content: strings.Join(matches, "\n")}, nil
}

// globRecursive handles ** patterns using filepath.WalkDir.
func globRecursive(root, pattern string, limit int) []string {
	// Extract the file extension or suffix from the pattern after **
	suffix := ""
	if idx := strings.Index(pattern, "**"); idx >= 0 {
		suffix = pattern[idx+2:] // everything after **
		if strings.HasPrefix(suffix, "/") {
			suffix = suffix[1:]
		}
	}

	var matches []string
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip .git directories
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		// Match against suffix pattern
		if suffix != "" {
			matched, _ := filepath.Match(suffix, d.Name())
			if !matched {
				return nil
			}
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		matches = append(matches, rel)
		if len(matches) >= limit {
			return fmt.Errorf("limit reached")
		}
		return nil
	})

	return matches
}
