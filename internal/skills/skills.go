// Package skills handles on-demand markdown skill loading.
// Skills are markdown files with YAML frontmatter, compatible with the anthropics/skills format.
package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Skill represents a loaded skill.
type Skill struct {
	Name        string
	Description string
	Content     string
	Source      string // origin of the skill (e.g. "user", "bundled")
	Path        string // absolute file path
}

// Loader reads and parses skill files from disk.
type Loader struct {
	dirs []string
}

// NewLoader creates a skill loader that searches the given directories.
func NewLoader(dirs ...string) *Loader {
	return &Loader{dirs: dirs}
}

// Load reads and parses a skill by name, searching each directory for <name>.md.
func (l *Loader) Load(ctx context.Context, name string) (*Skill, error) {
	filename := name
	if filepath.Ext(filename) != ".md" {
		filename += ".md"
	}

	for _, dir := range l.dirs {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("load skill %q: %w", name, ctx.Err())
		default:
		}

		path := filepath.Join(dir, filename)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("load skill %q: %w", name, err)
		}

		skillName, description, body := parseFrontmatter(name, string(data))
		return &Skill{
			Name:        skillName,
			Description: description,
			Content:     body,
			Source:      "file",
			Path:        path,
		}, nil
	}

	return nil, nil
}

// LoadByName reads and parses a skill by name. It is an explicit alias for Load.
func (l *Loader) LoadByName(ctx context.Context, name string) (*Skill, error) {
	return l.Load(ctx, name)
}

// LoadAll discovers and loads all *.md skill files from all configured directories,
// returning them sorted by name.
func (l *Loader) LoadAll(ctx context.Context) ([]*Skill, error) {
	var skills []*Skill
	seen := make(map[string]bool)

	for _, dir := range l.dirs {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("load all skills: %w", ctx.Err())
		default:
		}

		pattern := filepath.Join(dir, "*.md")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("glob skills in %s: %w", dir, err)
		}

		for _, path := range matches {
			if seen[path] {
				continue
			}
			seen[path] = true

			base := filepath.Base(path)
			defaultName := strings.TrimSuffix(base, filepath.Ext(base))

			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read skill %s: %w", path, err)
			}

			name, description, body := parseFrontmatter(defaultName, string(data))
			skills = append(skills, &Skill{
				Name:        name,
				Description: description,
				Content:     body,
				Source:      "file",
				Path:        path,
			})
		}
	}

	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})

	return skills, nil
}
