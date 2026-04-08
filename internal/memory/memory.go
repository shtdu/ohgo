// Package memory implements persistent cross-session memory.
package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Store manages memory files for a project.
type Store struct {
	dir string // project memory directory
	mu  sync.Mutex
}

// NewStore creates a memory store for the given working directory.
func NewStore(cwd string) (*Store, error) {
	dir, err := ProjectDir(cwd)
	if err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

// List returns sorted .md file basenames in the memory directory.
func (s *Store) List() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" && e.Name() != "MEMORY.md" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

var slugRe = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// Add creates a memory file and appends it to the MEMORY.md index.
func (s *Store) Add(title, content string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	slug := slugRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(title)), "_")
	slug = strings.Trim(slug, "_")
	if slug == "" {
		slug = "memory"
	}

	filename := slug + ".md"
	path := filepath.Join(s.dir, filename)
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		return "", fmt.Errorf("write memory file: %w", err)
	}

	entrypoint := filepath.Join(s.dir, "MEMORY.md")
	existing := "# Memory Index\n"
	if data, err := os.ReadFile(entrypoint); err == nil {
		existing = string(data)
	}

	if !strings.Contains(existing, filename) {
		existing = strings.TrimRight(existing, "\n") + fmt.Sprintf("\n- [%s](%s)\n", title, filename)
		if err := os.WriteFile(entrypoint, []byte(existing), 0o644); err != nil {
			return "", fmt.Errorf("update memory index: %w", err)
		}
	}

	return path, nil
}

// Remove deletes a memory file and removes its entry from the MEMORY.md index.
func (s *Store) Remove(name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Find file by stem or full name.
	var target string
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		stem := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		if stem == name || e.Name() == name {
			target = e.Name()
			break
		}
	}
	if target == "" {
		return false, nil
	}

	if err := os.Remove(filepath.Join(s.dir, target)); err != nil {
		return false, fmt.Errorf("remove memory file: %w", err)
	}

	entrypoint := filepath.Join(s.dir, "MEMORY.md")
	data, err := os.ReadFile(entrypoint)
	if err != nil {
		return true, nil
	}

	var kept []string
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.Contains(line, target) {
			kept = append(kept, line)
		}
	}
	os.WriteFile(entrypoint, []byte(strings.TrimRight(strings.Join(kept, "\n"), "\n")+"\n"), 0o644)

	return true, nil
}

// LoadPrompt reads the MEMORY.md index for prompt injection.
func (s *Store) LoadPrompt(maxLines int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entrypoint := filepath.Join(s.dir, "MEMORY.md")
	data, err := os.ReadFile(entrypoint)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	content := string(data)
	if maxLines <= 0 {
		return content, nil
	}

	lines := strings.Split(content, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n"), nil
}
