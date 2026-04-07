package hooks

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeHooksJSON(t *testing.T, dir string, manifest HookManifest) {
	t.Helper()
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hooks.json"), data, 0644))
}

func TestReloaderInitialLoad(t *testing.T) {
	dir := t.TempDir()
	writeHooksJSON(t, dir, HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo initial"}`),
		},
	})

	rl, err := NewReloader(dir)
	require.NoError(t, err)

	hooks := rl.Registry().Get(HookEventPreToolUse)
	require.Len(t, hooks, 1)
	assert.Equal(t, "echo initial", hooks[0].Command)
}

func TestReloaderModifyTriggersReload(t *testing.T) {
	dir := t.TempDir()
	writeHooksJSON(t, dir, HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo old"}`),
		},
	})

	rl, err := NewReloader(dir)
	require.NoError(t, err)
	assert.Equal(t, "echo old", rl.Registry().Get(HookEventPreToolUse)[0].Command)

	// Overwrite with new content and bump mod time
	time.Sleep(10 * time.Millisecond) // ensure distinct mtime
	writeHooksJSON(t, dir, HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo new"}`),
		},
	})
	// Force a distinct mod time to avoid filesystem resolution issues
	newTime := time.Now().Add(1 * time.Second)
	require.NoError(t, os.Chtimes(filepath.Join(dir, "hooks.json"), newTime, newTime))

	err = rl.CheckAndReload(context.Background())
	require.NoError(t, err)

	hooks := rl.Registry().Get(HookEventPreToolUse)
	require.Len(t, hooks, 1)
	assert.Equal(t, "echo new", hooks[0].Command)
}

func TestReloaderNoChangeNoReload(t *testing.T) {
	dir := t.TempDir()
	writeHooksJSON(t, dir, HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo stable"}`),
		},
	})

	rl, err := NewReloader(dir)
	require.NoError(t, err)

	regBefore := rl.Registry()

	err = rl.CheckAndReload(context.Background())
	require.NoError(t, err)

	regAfter := rl.Registry()
	// Same pointer means no reload occurred
	assert.Equal(t, regBefore, regAfter)
}

func TestReloaderDeleteHooksFile(t *testing.T) {
	dir := t.TempDir()
	writeHooksJSON(t, dir, HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo delete-me"}`),
		},
	})

	rl, err := NewReloader(dir)
	require.NoError(t, err)
	assert.Len(t, rl.Registry().Get(HookEventPreToolUse), 1)

	// Delete the hooks file
	require.NoError(t, os.Remove(filepath.Join(dir, "hooks.json")))

	err = rl.CheckAndReload(context.Background())
	require.NoError(t, err)

	assert.Empty(t, rl.Registry().Get(HookEventPreToolUse))
	assert.Empty(t, rl.Registry().Events())
}

func TestReloaderWatchAndReloadFires(t *testing.T) {
	dir := t.TempDir()
	writeHooksJSON(t, dir, HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo before-watch"}`),
		},
	})

	rl, err := NewReloader(dir)
	require.NoError(t, err)

	stop := rl.WatchAndReload(context.Background(), 100*time.Millisecond)
	defer stop()

	// Update the hooks file with a distinct mtime
	time.Sleep(50 * time.Millisecond)
	writeHooksJSON(t, dir, HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo after-watch"}`),
		},
	})
	newTime := time.Now().Add(1 * time.Second)
	require.NoError(t, os.Chtimes(filepath.Join(dir, "hooks.json"), newTime, newTime))

	// Wait for the watcher to pick it up
	time.Sleep(300 * time.Millisecond)

	hooks := rl.Registry().Get(HookEventPreToolUse)
	require.Len(t, hooks, 1)
	assert.Equal(t, "echo after-watch", hooks[0].Command)
}

func TestReloaderContextCancellationStopsWatcher(t *testing.T) {
	dir := t.TempDir()
	writeHooksJSON(t, dir, HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo ctx"}`),
		},
	})

	rl, err := NewReloader(dir)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	stop := rl.WatchAndReload(ctx, 50*time.Millisecond)

	// Cancel context and then call stop — stop should return promptly
	cancel()

	done := make(chan struct{})
	go func() {
		stop()
		close(done)
	}()

	select {
	case <-done:
		// Success: stop returned after cancellation
	case <-time.After(2 * time.Second):
		t.Fatal("stop() did not return after context cancellation")
	}
}

func TestReloaderConcurrentRegistryAccess(t *testing.T) {
	dir := t.TempDir()
	writeHooksJSON(t, dir, HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo concurrent"}`),
		},
	})

	rl, err := NewReloader(dir)
	require.NoError(t, err)

	var wg sync.WaitGroup

	// Concurrent readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = rl.Registry()
			_ = rl.Registry().Get(HookEventPreToolUse)
		}()
	}

	// Concurrent reloads (no file changes, so no actual reload)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = rl.CheckAndReload(context.Background())
		}()
	}

	wg.Wait()

	// Registry should still be valid
	hooks := rl.Registry().Get(HookEventPreToolUse)
	require.Len(t, hooks, 1)
	assert.Equal(t, "echo concurrent", hooks[0].Command)
}

func TestReloaderMissingDirectory(t *testing.T) {
	rl, err := NewReloader("/nonexistent/path/that/does/not/exist")
	require.NoError(t, err)
	require.NotNil(t, rl)

	// Should have an empty registry
	assert.Empty(t, rl.Registry().Events())
	assert.Equal(t, "no hooks registered", rl.Registry().Summary())
}
