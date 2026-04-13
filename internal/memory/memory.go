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

// Store manages memory files across personal and project layers.
// Personal memory is stored in ~/.ohgo/data/memory/_personal/.
// Project memory is stored per-project under ~/.ohgo/data/memory/<project>-<hash>/.
// Each entry is a markdown file; MEMORY.md serves as the index.
// Format is compatible with the Python OpenHarness version.
// All operations are safe for concurrent use.
type Store struct {
	projectDir  string // project memory directory
	personalDir string // personal (user-level) memory directory
	mu          sync.Mutex
}

// NewStore creates a memory store for the given working directory.
// Both personal and project memory directories are initialized.
func NewStore(cwd string) (*Store, error) {
	projDir, err := ProjectDir(cwd)
	if err != nil {
		return nil, err
	}
	persDir, err := PersonalDir()
	if err != nil {
		return nil, err
	}
	return &Store{projectDir: projDir, personalDir: persDir}, nil
}

// ProjectDir returns the project-scoped memory directory path.
func (s *Store) ProjectDir() string { return s.projectDir }

// PersonalDir returns the user-level memory directory path.
func (s *Store) PersonalDir() string { return s.personalDir }

// --- Project-scoped operations (existing API preserved) ---

// List returns sorted .md file basenames in the project memory directory.
func (s *Store) List() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return listDir(s.projectDir)
}

// Add creates a project memory file and appends it to the project MEMORY.md index.
func (s *Store) Add(title, content string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return addToDir(s.projectDir, title, content)
}

// Remove deletes a project memory file and removes its entry from the index.
func (s *Store) Remove(name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return removeFromDir(s.projectDir, name)
}

// --- Personal-scoped operations ---

// ListPersonal returns sorted .md file basenames in the personal memory directory.
func (s *Store) ListPersonal() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return listDir(s.personalDir)
}

// AddPersonal creates a personal memory file and appends it to the personal MEMORY.md index.
func (s *Store) AddPersonal(title, content string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return addToDir(s.personalDir, title, content)
}

// RemovePersonal deletes a personal memory file and removes its entry from the index.
func (s *Store) RemovePersonal(name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return removeFromDir(s.personalDir, name)
}

// --- Dual-layer operations ---

// LoadPrompt reads both personal and project MEMORY.md indexes for prompt injection.
// Personal memory appears first, then project memory, separated by section headers.
func (s *Store) LoadPrompt(maxLines int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	personalContent, err := readIndex(s.personalDir)
	if err != nil {
		return "", fmt.Errorf("read personal memory: %w", err)
	}
	projectContent, err := readIndex(s.projectDir)
	if err != nil {
		return "", fmt.Errorf("read project memory: %w", err)
	}

	if personalContent == "" && projectContent == "" {
		return "", nil
	}

	var sections []string
	if personalContent != "" {
		sections = append(sections, "# Personal Memory\n"+personalContent)
	}
	if projectContent != "" {
		sections = append(sections, "# Project Memory\n"+projectContent)
	}

	content := strings.Join(sections, "\n\n")

	if maxLines <= 0 {
		return content, nil
	}

	lines := strings.Split(content, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n"), nil
}

// --- Internal helpers ---

func readIndex(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "MEMORY.md"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func listDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
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

func addToDir(dir, title, content string) (string, error) {
	slug := slugRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(title)), "_")
	slug = strings.Trim(slug, "_")
	if slug == "" {
		slug = "memory"
	}

	filename := slug + ".md"
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		return "", fmt.Errorf("write memory file: %w", err)
	}

	entrypoint := filepath.Join(dir, "MEMORY.md")
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

func removeFromDir(dir, name string) (bool, error) {
	var target string
	entries, err := os.ReadDir(dir)
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

	if err := os.Remove(filepath.Join(dir, target)); err != nil {
		return false, fmt.Errorf("remove memory file: %w", err)
	}

	entrypoint := filepath.Join(dir, "MEMORY.md")
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
	if err := os.WriteFile(entrypoint, []byte(strings.TrimRight(strings.Join(kept, "\n"), "\n")+"\n"), 0o644); err != nil {
		return false, fmt.Errorf("write entrypoint: %w", err)
	}

	return true, nil
}
