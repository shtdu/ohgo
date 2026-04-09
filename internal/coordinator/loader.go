package coordinator

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader discovers and parses agent definition files from a set of directories.
type Loader struct {
	dirs []string
}

// NewLoader creates a Loader that searches the given directories for agent YAML files.
func NewLoader(dirs ...string) *Loader {
	return &Loader{dirs: dirs}
}

// LoadAll reads all *.yaml files from the configured directories and returns
// parsed AgentDefs. Invalid files are skipped with a log message.
func (l *Loader) LoadAll(ctx context.Context) ([]*AgentDef, error) {
	var defs []*AgentDef

	for _, dir := range l.dirs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read dir %s: %w", dir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
				continue
			}

			path := filepath.Join(dir, entry.Name())
			def, err := parseAgentFile(path)
			if err != nil {
				log.Printf("coordinator: skipping invalid agent file %s: %v", path, err)
				continue
			}
			if def.Name == "" {
				log.Printf("coordinator: skipping agent file %s: missing name", path)
				continue
			}
			defs = append(defs, def)
		}
	}

	return defs, nil
}

// parseAgentFile reads a single YAML file and returns the parsed AgentDef.
func parseAgentFile(path string) (*AgentDef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var def AgentDef
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return &def, nil
}
