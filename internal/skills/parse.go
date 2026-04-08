package skills

import (
	"fmt"
	"strings"
)

// parseFrontmatter extracts name, description, and body from a skill markdown file.
// It first tries YAML frontmatter (delimited by ---), then falls back to heading
// and first-paragraph extraction. This is a direct port of the Python
// _parse_skill_markdown function.
func parseFrontmatter(defaultName, content string) (name, description, body string) {
	name = defaultName
	description = ""
	lines := strings.Split(content, "\n")

	// Track where the body starts (after frontmatter, if present).
	bodyStart := 0

	// Try YAML frontmatter first (--- ... ---).
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				// Parse frontmatter fields between the two delimiters.
				for _, fmLine := range lines[1:i] {
					fmStripped := strings.TrimSpace(fmLine)
					if strings.HasPrefix(fmStripped, "name:") {
						val := strings.TrimSpace(fmStripped[5:])
						val = strings.Trim(val, `"'`)
						if val != "" {
							name = val
						}
					} else if strings.HasPrefix(fmStripped, "description:") {
						val := strings.TrimSpace(fmStripped[12:])
						val = strings.Trim(val, `"'`)
						if val != "" {
							description = val
						}
					}
				}
				bodyStart = i + 1
				break
			}
		}
	}

	// Fallback: extract from headings and first paragraph (after frontmatter).
	if description == "" {
		for _, line := range lines[bodyStart:] {
			stripped := strings.TrimSpace(line)
			if strings.HasPrefix(stripped, "# ") {
				if name == "" || name == defaultName {
					heading := strings.TrimSpace(stripped[2:])
					if heading != "" {
						name = heading
					}
				}
				continue
			}
			if stripped != "" && !strings.HasPrefix(stripped, "#") {
				if len(stripped) > 200 {
					description = stripped[:200]
				} else {
					description = stripped
				}
				break
			}
		}
	}

	if description == "" {
		description = fmt.Sprintf("Skill: %s", name)
	}

	// Body is everything after the frontmatter block.
	if bodyStart < len(lines) {
		body = strings.Join(lines[bodyStart:], "\n")
	}

	return name, description, body
}
