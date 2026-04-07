package hooks

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Reloader watches a directory for changes to hooks.json files
// and reloads the registry when modifications are detected.
type Reloader struct {
	mu       sync.RWMutex
	registry *Registry
	dir      string
	modTimes map[string]time.Time
}

// NewReloader creates a reloader that watches the given directory.
// It performs an initial load.
func NewReloader(dir string) (*Reloader, error) {
	r := &Reloader{
		dir:      dir,
		modTimes: make(map[string]time.Time),
	}

	ctx := context.Background()
	registry, err := LoadFromDir(ctx, dir)
	if err != nil {
		registry = NewRegistry() // start with empty on error
	}
	r.registry = registry
	r.scanModTimes()

	return r, nil
}

// Registry returns the current hook registry.
func (r *Reloader) Registry() *Registry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registry
}

// CheckAndReload checks if any hooks.json files have been modified
// since the last check, and reloads if so.
func (r *Reloader) CheckAndReload(ctx context.Context) error {
	currentModTimes := r.scanModTimesCurrent()

	changed := false
	// Check for modified or new files
	for path, modTime := range currentModTimes {
		if oldModTime, exists := r.modTimes[path]; !exists || !modTime.Equal(oldModTime) {
			changed = true
			break
		}
	}

	// Check for deleted files
	if !changed {
		for path := range r.modTimes {
			if _, exists := currentModTimes[path]; !exists {
				changed = true
				break
			}
		}
	}

	if !changed {
		return nil
	}

	registry, err := LoadFromDir(ctx, r.dir)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.registry = registry
	r.modTimes = currentModTimes
	r.mu.Unlock()

	return nil
}

// WatchAndReload starts a goroutine that periodically checks for changes.
// Returns a stop function.
func (r *Reloader) WatchAndReload(ctx context.Context, interval time.Duration) (stop func()) {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = r.CheckAndReload(ctx)
			}
		}
	}()

	return func() {
		cancel()
		<-done
	}
}

// scanModTimes populates the initial modTime cache.
func (r *Reloader) scanModTimes() {
	r.modTimes = r.scanModTimesCurrent()
}

func (r *Reloader) scanModTimesCurrent() map[string]time.Time {
	times := make(map[string]time.Time)
	filepath.WalkDir(r.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "hooks.json" {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		times[path] = info.ModTime()
		return nil
	})
	return times
}
