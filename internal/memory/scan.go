package memory

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Scan reads memory files from the project memory directory and returns
// headers sorted by modification time (newest first), capped at maxFiles.
func Scan(cwd string, maxFiles int) ([]*Header, error) {
	dir, err := ProjectDir(cwd)
	if err != nil {
		return nil, err
	}
	return scanDir(dir, maxFiles)
}

// scanDir reads memory files from the given directory and returns
// headers sorted by modification time (newest first), capped at maxFiles.
func scanDir(dir string, maxFiles int) ([]*Header, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var headers []*Header
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		if entry.Name() == "MEMORY.md" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		headers = append(headers, parseFile(path, string(data), info.ModTime()))
	}

	sort.Slice(headers, func(i, j int) bool {
		return headers[i].ModifiedAt.After(headers[j].ModifiedAt)
	})

	if maxFiles > 0 && len(headers) > maxFiles {
		headers = headers[:maxFiles]
	}
	return headers, nil
}

// parseFile extracts a Header from a memory file's content.
func parseFile(path, content string, modTime time.Time) *Header {
	lines := strings.Split(content, "\n")
	title := strings.TrimSuffix(filepath.Base(path), filepath.Ext(filepath.Base(path)))
	description := ""
	memoryType := ""
	bodyStart := 0

	// Parse YAML frontmatter (--- ... ---).
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				for _, fmLine := range lines[1:i] {
					key, val, ok := parseKV(fmLine)
					if !ok {
						continue
					}
					switch key {
					case "name":
						title = val
					case "description":
						description = val
					case "type":
						memoryType = val
					}
				}
				bodyStart = i + 1
				break
			}
		}
	}

	// Fallback: first non-empty, non-frontmatter, non-heading line as description.
	descLineIdx := -1
	if description == "" {
		for idx := bodyStart; idx < len(lines) && idx < bodyStart+10; idx++ {
			stripped := strings.TrimSpace(lines[idx])
			if stripped != "" && stripped != "---" && !strings.HasPrefix(stripped, "#") {
				if len(stripped) > 200 {
					description = stripped[:200]
				} else {
					description = stripped
				}
				descLineIdx = idx
				break
			}
		}
	}

	// Build body preview.
	var bodyLines []string
	for idx := bodyStart; idx < len(lines); idx++ {
		stripped := strings.TrimSpace(lines[idx])
		if stripped == "" || strings.HasPrefix(stripped, "#") || idx == descLineIdx {
			continue
		}
		bodyLines = append(bodyLines, stripped)
	}
	bodyPreview := strings.Join(bodyLines, " ")
	if len(bodyPreview) > 300 {
		bodyPreview = bodyPreview[:300]
	}

	return &Header{
		Path:        path,
		Title:       title,
		Description: description,
		ModifiedAt:  modTime,
		MemoryType:  memoryType,
		BodyPreview: bodyPreview,
	}
}

// parseKV extracts a key-value pair from a frontmatter line like "name: value".
func parseKV(line string) (key, value string, ok bool) {
	stripped := strings.TrimSpace(line)
	key, val, found := strings.Cut(stripped, ":")
	if !found {
		return "", "", false
	}
	key = strings.TrimSpace(key)
	val = strings.TrimSpace(val)
	val = strings.Trim(val, `"'`)
	if val == "" {
		return key, val, false
	}
	return key, val, true
}
