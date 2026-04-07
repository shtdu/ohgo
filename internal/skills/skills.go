// Package skills handles on-demand markdown skill loading.
// Skills are markdown files with YAML frontmatter, compatible with the anthropics/skills format.
package skills

import (
	"context"
)

// Skill represents a loaded skill.
type Skill struct {
	Name        string
	Description string
	Content     string
}

// Loader reads and parses skill files from disk.
type Loader struct {
	dirs []string
}

// NewLoader creates a skill loader that searches the given directories.
func NewLoader(dirs ...string) *Loader {
	return &Loader{dirs: dirs}
}

// Load reads and parses a skill by name.
func (l *Loader) Load(ctx context.Context, name string) (*Skill, error) {
	// TODO: implement skill discovery and parsing
	return nil, nil
}
