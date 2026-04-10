// Package grep implements the grep tool for content search.
package grep

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

const (
	defaultGrepLimit = 200
	maxGrepLimit     = 2000
	binaryCheckSize  = 8192
)

type grepInput struct {
	Pattern       string `json:"pattern"`
	Path          string `json:"path"`
	Glob          string `json:"glob"`
	CaseSensitive bool   `json:"case_sensitive"`
	Limit         int    `json:"limit"`
}

type match struct {
	filePath string
	lineNum  int
	line     string
}

// GrepTool searches file contents using regular expressions.
type GrepTool struct{}

func (GrepTool) Name() string { return "grep" }

func (GrepTool) Description() string {
	return "Search file contents using regular expressions. Supports case-insensitive search and file filtering."
}

func (GrepTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pattern": map[string]any{
				"type":        "string",
				"description": "Regular expression pattern to search for",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "File or directory to search in",
				"default":     ".",
			},
			"glob": map[string]any{
				"type":        "string",
				"description": "File pattern to filter (e.g. '*.go', '*.ts')",
				"default":     "**/*",
			},
			"case_sensitive": map[string]any{
				"type":        "boolean",
				"description": "Whether search is case sensitive",
				"default":     true,
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of matches",
				"default":     defaultGrepLimit,
				"minimum":     1,
				"maximum":     maxGrepLimit,
			},
		},
		"required":             []string{"pattern"},
		"additionalProperties": false,
	}
}

func (GrepTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input grepInput
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

	// Compile regex
	flags := ""
	if !input.CaseSensitive {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + input.Pattern)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid regex: %v", err), IsError: true}, nil
	}

	// Defaults
	searchPath := input.Path
	if searchPath == "" {
		searchPath = "."
	}
	limit := input.Limit
	if limit <= 0 {
		limit = defaultGrepLimit
	}
	limit = min(limit, maxGrepLimit)

	cwd, err := os.Getwd()
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("cannot get working directory: %v", err), IsError: true}, nil
	}

	var matches []match

	info, err := os.Stat(searchPath)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("path not found: %s", searchPath), IsError: true}, nil
	}

	if !info.IsDir() {
		// Search single file
		fileMatches := searchFile(searchPath, re, cwd, limit)
		matches = append(matches, fileMatches...)
	} else {
		// Walk directory
		_ = filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if d.IsDir() {
				if d.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}

			// Apply glob filter against relative path from search root
			if input.Glob != "" && input.Glob != "**/*" && input.Glob != "*" {
				rel, relErr := filepath.Rel(searchPath, path)
				if relErr != nil {
					rel = d.Name()
				}
				if !globMatch(input.Glob, rel) {
					return nil
				}
			}

			fileMatches := searchFile(path, re, cwd, limit-len(matches))
			matches = append(matches, fileMatches...)
			if len(matches) >= limit {
				return fmt.Errorf("limit reached")
			}
			return nil
		})
	}

	if len(matches) == 0 {
		return tools.Result{Content: "(no matches)"}, nil
	}

	// Sort by file path then line number
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].filePath != matches[j].filePath {
			return matches[i].filePath < matches[j].filePath
		}
		return matches[i].lineNum < matches[j].lineNum
	})

	// Truncate if over limit
	if len(matches) > limit {
		matches = matches[:limit]
	}

	var buf strings.Builder
	for _, m := range matches {
		fmt.Fprintf(&buf, "%s:%d:%s\n", m.filePath, m.lineNum, m.line)
	}

	return tools.Result{Content: strings.TrimRight(buf.String(), "\n")}, nil
}

func searchFile(path string, re *regexp.Regexp, cwd string, limit int) []match {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	// Binary check
	buf := make([]byte, binaryCheckSize)
	n, _ := f.Read(buf)
	if n > 0 && bytes.ContainsRune(buf[:n], 0) {
		return nil
	}
	_, _ = f.Seek(0, 0)

	relPath, err := filepath.Rel(cwd, path)
	if err != nil {
		relPath = path
	}

	var matches []match
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if re.MatchString(scanner.Text()) {
			matches = append(matches, match{filePath: relPath, lineNum: lineNum, line: scanner.Text()})
			if len(matches) >= limit {
				break
			}
		}
	}
	return matches
}

// globMatch does simple glob matching for file name filtering.
func globMatch(pattern, name string) bool {
	matched, _ := filepath.Match(pattern, name)
	return matched
}
