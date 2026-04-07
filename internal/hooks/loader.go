package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Registry stores hooks grouped by event type.
type Registry struct {
	mu    sync.RWMutex
	hooks map[HookEvent][]HookDefinition
}

// NewRegistry creates an empty hook registry.
func NewRegistry() *Registry {
	return &Registry{
		hooks: make(map[HookEvent][]HookDefinition),
	}
}

// Register adds a hook definition for a given event.
func (r *Registry) Register(event HookEvent, hook HookDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks[event] = append(r.hooks[event], hook)
}

// Get returns all hooks registered for an event.
func (r *Registry) Get(event HookEvent) []HookDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]HookDefinition(nil), r.hooks[event]...)
}

// Events returns all events that have registered hooks.
func (r *Registry) Events() []HookEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var events []HookEvent
	for e := range r.hooks {
		events = append(events, e)
	}
	return events
}

// Summary returns a human-readable description of all registered hooks.
func (r *Registry) Summary() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.hooks) == 0 {
		return "no hooks registered"
	}
	var summary strings.Builder
	for event, hooks := range r.hooks {
		fmt.Fprintf(&summary, "%s: %d hook(s)\n", event, len(hooks))
	}
	return summary.String()
}

// LoadFromDir reads all hooks.json files found under the given directory tree.
// Returns a Registry with all discovered hooks. If the directory doesn't exist,
// returns an empty registry with no error.
func LoadFromDir(ctx context.Context, dir string) (*Registry, error) {
	registry := NewRegistry()

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return registry, nil
	}

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors, continue walking
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.IsDir() || d.Name() != "hooks.json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		var manifest HookManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		for _, raw := range manifest.Hooks {
			var hook HookDefinition
			if err := json.Unmarshal(raw, &hook); err != nil {
				continue // skip malformed hooks
			}
			if err := hook.Validate(); err != nil {
				continue // skip invalid hooks
			}
			registry.Register(hook.Event, hook)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return registry, nil
}
