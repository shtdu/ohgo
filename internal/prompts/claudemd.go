package prompts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CLAUDEmdFile represents a discovered instruction file.
type CLAUDEmdFile struct {
	Path    string
	Content string
}

// DiscoverCLAUDEmd walks from cwd upward, collecting CLAUDE.md files
// and .claude/rules/*.md files. Returns files in discovery order
// (closest to cwd first).
func DiscoverCLAUDEmd(ctx context.Context, cwd string) ([]CLAUDEmdFile, error) {
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return nil, fmt.Errorf("resolve cwd: %w", err)
	}

	var files []CLAUDEmdFile
	seen := make(map[string]bool)
	dir := absCwd

	for {
		select {
		case <-ctx.Done():
			return files, ctx.Err()
		default:
		}

		// Check CLAUDE.md
		claudeMdPath := filepath.Join(dir, "CLAUDE.md")
		if content, err := os.ReadFile(claudeMdPath); err == nil {
			abs, _ := filepath.Abs(claudeMdPath)
			if !seen[abs] {
				seen[abs] = true
				files = append(files, CLAUDEmdFile{Path: claudeMdPath, Content: string(content)})
			}
		}

		// Check .claude/CLAUDE.md
		dotClaudeMd := filepath.Join(dir, ".claude", "CLAUDE.md")
		if content, err := os.ReadFile(dotClaudeMd); err == nil {
			abs, _ := filepath.Abs(dotClaudeMd)
			if !seen[abs] {
				seen[abs] = true
				files = append(files, CLAUDEmdFile{Path: dotClaudeMd, Content: string(content)})
			}
		}

		// Check .claude/rules/*.md
		rulesDir := filepath.Join(dir, ".claude", "rules")
		if entries, err := os.ReadDir(rulesDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				rulePath := filepath.Join(rulesDir, entry.Name())
				abs, _ := filepath.Abs(rulePath)
				if seen[abs] {
					continue
				}
				if content, err := os.ReadFile(rulePath); err == nil {
					seen[abs] = true
					files = append(files, CLAUDEmdFile{Path: rulePath, Content: string(content)})
				}
			}
		}

		// Walk up
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}

	return files, nil
}

// MergeCLAUDEmd combines discovered files into a single prompt section.
// Each file is truncated to maxCharsPerFile.
// Returns nil if no files found.
func MergeCLAUDEmd(files []CLAUDEmdFile, maxCharsPerFile int) *string {
	if len(files) == 0 {
		return nil
	}

	var sb strings.Builder
	for i, f := range files {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		content := f.Content
		if len(content) > maxCharsPerFile {
			content = content[:maxCharsPerFile] + "\n... (truncated)"
		}
		fmt.Fprintf(&sb, "# Project instructions from %s\n\n%s", f.Path, content)
	}

	result := sb.String()
	return &result
}
