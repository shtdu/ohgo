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

// --- Registry tests ---

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	require.NotNil(t, r)
	assert.Empty(t, r.Events())
	assert.Equal(t, "no hooks registered", r.Summary())
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	hook := HookDefinition{
		Event:   HookEventPreToolUse,
		Type:    HookTypeCommand,
		Matcher: "bash",
		Command: "echo check",
	}
	r.Register(HookEventPreToolUse, hook)

	got := r.Get(HookEventPreToolUse)
	require.Len(t, got, 1)
	assert.Equal(t, "bash", got[0].Matcher)
	assert.Equal(t, "echo check", got[0].Command)

	// Getting a different event should return empty
	assert.Empty(t, r.Get(HookEventPostToolUse))
}

func TestRegistryGetReturnsCopy(t *testing.T) {
	r := NewRegistry()
	hook := HookDefinition{
		Event:   HookEventPreToolUse,
		Type:    HookTypeCommand,
		Command: "echo hi",
	}
	r.Register(HookEventPreToolUse, hook)

	slice := r.Get(HookEventPreToolUse)
	slice[0].Command = "mutated"

	// Original should be unchanged
	original := r.Get(HookEventPreToolUse)
	assert.Equal(t, "echo hi", original[0].Command)
}

func TestRegistryMultipleHooksPerEvent(t *testing.T) {
	r := NewRegistry()
	r.Register(HookEventPreToolUse, HookDefinition{
		Event: HookEventPreToolUse, Type: HookTypeCommand, Command: "first",
	})
	r.Register(HookEventPreToolUse, HookDefinition{
		Event: HookEventPreToolUse, Type: HookTypeCommand, Command: "second",
	})
	r.Register(HookEventPostToolUse, HookDefinition{
		Event: HookEventPostToolUse, Type: HookTypeHTTP, URL: "http://localhost",
	})

	assert.Len(t, r.Get(HookEventPreToolUse), 2)
	assert.Len(t, r.Get(HookEventPostToolUse), 1)
}

func TestRegistryEvents(t *testing.T) {
	r := NewRegistry()
	r.Register(HookEventPreToolUse, HookDefinition{
		Event: HookEventPreToolUse, Type: HookTypeCommand, Command: "echo",
	})
	r.Register(HookEventPostToolUse, HookDefinition{
		Event: HookEventPostToolUse, Type: HookTypeHTTP, URL: "http://localhost",
	})

	events := r.Events()
	assert.Len(t, events, 2)
	assert.Contains(t, events, HookEventPreToolUse)
	assert.Contains(t, events, HookEventPostToolUse)
}

func TestRegistrySummary(t *testing.T) {
	r := NewRegistry()
	assert.Equal(t, "no hooks registered", r.Summary())

	r.Register(HookEventPreToolUse, HookDefinition{
		Event: HookEventPreToolUse, Type: HookTypeCommand, Command: "echo",
	})
	r.Register(HookEventPostToolUse, HookDefinition{
		Event: HookEventPostToolUse, Type: HookTypeHTTP, URL: "http://localhost",
	})

	summary := r.Summary()
	assert.Contains(t, summary, "pre_tool_use: 1 hook(s)")
	assert.Contains(t, summary, "post_tool_use: 1 hook(s)")
}

func TestRegistryConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r.Register(HookEventPreToolUse, HookDefinition{
				Event:   HookEventPreToolUse,
				Type:    HookTypeCommand,
				Command: "cmd",
			})
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.Get(HookEventPreToolUse)
			_ = r.Events()
			_ = r.Summary()
		}()
	}

	wg.Wait()

	got := r.Get(HookEventPreToolUse)
	assert.Len(t, got, 50)
}

// --- LoadFromDir tests ---

func TestLoadFromDirHappyPath(t *testing.T) {
	dir := t.TempDir()
	manifest := HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "matcher": "bash", "command": "echo check"}`),
			[]byte(`{"event": "post_tool_use", "type": "http", "url": "http://localhost:8080/hook"}`),
		},
	}
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hooks.json"), data, 0644))

	reg, err := LoadFromDir(context.Background(), dir)
	require.NoError(t, err)

	preHooks := reg.Get(HookEventPreToolUse)
	require.Len(t, preHooks, 1)
	assert.Equal(t, HookTypeCommand, preHooks[0].Type)
	assert.Equal(t, "bash", preHooks[0].Matcher)
	assert.Equal(t, "echo check", preHooks[0].Command)

	postHooks := reg.Get(HookEventPostToolUse)
	require.Len(t, postHooks, 1)
	assert.Equal(t, HookTypeHTTP, postHooks[0].Type)
	assert.Equal(t, "http://localhost:8080/hook", postHooks[0].URL)
}

func TestLoadFromDirMissingDirectory(t *testing.T) {
	reg, err := LoadFromDir(context.Background(), "/nonexistent/path/that/does/not/exist")
	require.NoError(t, err)
	assert.Empty(t, reg.Events())
	assert.Equal(t, "no hooks registered", reg.Summary())
}

func TestLoadFromDirMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hooks.json"), []byte("{not valid json}"), 0644))

	_, err := LoadFromDir(context.Background(), dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestLoadFromDirNestedDirectories(t *testing.T) {
	dir := t.TempDir()

	// Create nested subdirectory structure
	subDir := filepath.Join(dir, "subdir", "nested")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	rootManifest := HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "root-cmd"}`),
		},
	}
	nestedManifest := HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "post_tool_use", "type": "http", "url": "http://nested/hook"}`),
		},
	}

	rootData, err := json.Marshal(rootManifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hooks.json"), rootData, 0644))

	nestedData, err := json.Marshal(nestedManifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "hooks.json"), nestedData, 0644))

	reg, err := LoadFromDir(context.Background(), dir)
	require.NoError(t, err)

	assert.Len(t, reg.Get(HookEventPreToolUse), 1)
	assert.Len(t, reg.Get(HookEventPostToolUse), 1)
}

func TestLoadFromDirContextCancellation(t *testing.T) {
	dir := t.TempDir()

	// Create a hooks.json so there's something to walk over
	manifest := HookManifest{
		Hooks: []json.RawMessage{
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo"}`),
		},
	}
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hooks.json"), data, 0644))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err = LoadFromDir(ctx, dir)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestLoadFromDirInvalidHookEntries(t *testing.T) {
	dir := t.TempDir()

	manifest := HookManifest{
		Hooks: []json.RawMessage{
			// Valid command hook
			[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo ok"}`),
			// Missing event — invalid, should be silently skipped
			[]byte(`{"type": "command", "command": "echo no-event"}`),
			// Missing command — invalid, should be silently skipped
			[]byte(`{"event": "pre_tool_use", "type": "command"}`),
			// Missing URL — invalid, should be silently skipped
			[]byte(`{"event": "post_tool_use", "type": "http"}`),
			// Not even a JSON object — malformed, should be silently skipped
			[]byte(`"just a string"`),
			// Valid HTTP hook
			[]byte(`{"event": "post_tool_use", "type": "http", "url": "http://valid/hook"}`),
		},
	}
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hooks.json"), data, 0644))

	reg, err := LoadFromDir(context.Background(), dir)
	require.NoError(t, err)

	// Only the two valid hooks should be registered
	assert.Len(t, reg.Get(HookEventPreToolUse), 1)
	assert.Len(t, reg.Get(HookEventPostToolUse), 1)
}

func TestLoadFromDirEmptyDirectory(t *testing.T) {
	dir := t.TempDir() // empty directory

	reg, err := LoadFromDir(context.Background(), dir)
	require.NoError(t, err)
	assert.Empty(t, reg.Events())
}

func TestLoadFromDirNoHooksFile(t *testing.T) {
	dir := t.TempDir()
	// Create a non-hooks.json file
	require.NoError(t, os.WriteFile(filepath.Join(dir, "other.json"), []byte(`{}`), 0644))

	reg, err := LoadFromDir(context.Background(), dir)
	require.NoError(t, err)
	assert.Empty(t, reg.Events())
}

func TestLoadFromDirTimeout(t *testing.T) {
	dir := t.TempDir()

	// Create many nested directories with hooks.json files
	for i := 0; i < 10; i++ {
		sub := filepath.Join(dir, "level", "deep")
		require.NoError(t, os.MkdirAll(sub, 0755))
		manifest := HookManifest{
			Hooks: []json.RawMessage{
				[]byte(`{"event": "pre_tool_use", "type": "command", "command": "echo"}`),
			},
		}
		data, err := json.Marshal(manifest)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(sub, "hooks.json"), data, 0644))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// The context will expire before or during the walk
	_, err := LoadFromDir(ctx, dir)
	if err != nil {
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	}
	// It's also acceptable for the walk to complete within the nanosecond,
	// so we don't assert that an error must occur.
}
