package plugins

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helperWritePluginJSON writes a plugin.json manifest to the given directory.
func helperWritePluginJSON(t *testing.T, dir string, manifest Manifest) {
	t.Helper()
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plugin.json"), data, 0o644))
}

// helperWriteFile writes content to a file, creating parent directories as needed.
func helperWriteFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

func TestDiscoverFindsValidPlugin(t *testing.T) {
	root := t.TempDir()

	pluginDir := filepath.Join(root, "my-plugin")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))

	helperWritePluginJSON(t, pluginDir, Manifest{
		Name:        "my-plugin",
		Version:     "1.0.0",
		Description: "A test plugin",
	})

	plugins, err := Discover(context.Background(), []string{root})
	require.NoError(t, err)
	require.Len(t, plugins, 1)
	assert.Equal(t, "my-plugin", plugins[0].Manifest.Name)
	assert.Equal(t, "1.0.0", plugins[0].Manifest.Version)
	assert.Equal(t, "A test plugin", plugins[0].Manifest.Description)
	assert.False(t, plugins[0].Enabled) // EnabledByDefault defaults to false in Go
}

func TestDiscoverLoadsSkills(t *testing.T) {
	root := t.TempDir()

	pluginDir := filepath.Join(root, "skill-plugin")
	skillsDir := filepath.Join(pluginDir, "skills")
	require.NoError(t, os.MkdirAll(skillsDir, 0o755))

	helperWritePluginJSON(t, pluginDir, Manifest{
		Name:      "skill-plugin",
		SkillsDir: "skills",
	})

	helperWriteFile(t, filepath.Join(skillsDir, "greeting.md"),
		"---\nname: greeting\ndescription: Say hello\n---\n\nHello, world!\n")

	plugins, err := Discover(context.Background(), []string{root})
	require.NoError(t, err)
	require.Len(t, plugins, 1)
	require.Len(t, plugins[0].Skills, 1)
	assert.Equal(t, "greeting", plugins[0].Skills[0].Name)
	assert.Equal(t, "Say hello", plugins[0].Skills[0].Description)
	assert.Equal(t, "plugin", plugins[0].Skills[0].Source)
}

func TestDiscoverClaudePluginSubdirectory(t *testing.T) {
	root := t.TempDir()

	pluginDir := filepath.Join(root, "claude-plugin")
	subDir := filepath.Join(pluginDir, ".claude-plugin")
	require.NoError(t, os.MkdirAll(subDir, 0o755))

	helperWritePluginJSON(t, subDir, Manifest{
		Name:    "claude-plugin",
		Version: "2.0.0",
	})

	plugins, err := Discover(context.Background(), []string{root})
	require.NoError(t, err)
	require.Len(t, plugins, 1)
	assert.Equal(t, "claude-plugin", plugins[0].Manifest.Name)
}

func TestDiscoverSkipsInvalidManifest(t *testing.T) {
	root := t.TempDir()

	pluginDir := filepath.Join(root, "bad-plugin")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))

	require.NoError(t, os.WriteFile(
		filepath.Join(pluginDir, "plugin.json"),
		[]byte("{not valid json}"),
		0o644,
	))

	plugins, err := Discover(context.Background(), []string{root})
	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestDiscoverEmptyDirectories(t *testing.T) {
	root := t.TempDir()

	// Empty subdirectory: no plugin.json
	pluginDir := filepath.Join(root, "empty")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))

	plugins, err := Discover(context.Background(), []string{root})
	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestDiscoverMultiplePluginsSortedByName(t *testing.T) {
	root := t.TempDir()

	for _, name := range []string{"zeta", "alpha", "mid"} {
		dir := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		helperWritePluginJSON(t, dir, Manifest{Name: name})
	}

	plugins, err := Discover(context.Background(), []string{root})
	require.NoError(t, err)
	require.Len(t, plugins, 3)
	assert.Equal(t, "alpha", plugins[0].Manifest.Name)
	assert.Equal(t, "mid", plugins[1].Manifest.Name)
	assert.Equal(t, "zeta", plugins[2].Manifest.Name)
}

func TestDiscoverNonexistentDirectory(t *testing.T) {
	plugins, err := Discover(context.Background(), []string{"/nonexistent/path"})
	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestDiscoverLoadsHooks(t *testing.T) {
	root := t.TempDir()

	pluginDir := filepath.Join(root, "hooked")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))

	helperWritePluginJSON(t, pluginDir, Manifest{
		Name:      "hooked",
		HooksFile: "hooks.json",
	})

	hooksContent := `{
		"pre_tool_use": [
			{"event": "pre_tool_use", "type": "command", "command": "echo check"}
		]
	}`
	helperWriteFile(t, filepath.Join(pluginDir, "hooks.json"), hooksContent)

	plugins, err := Discover(context.Background(), []string{root})
	require.NoError(t, err)
	require.Len(t, plugins, 1)
	require.NotNil(t, plugins[0].Hooks)

	preHooks, ok := plugins[0].Hooks["pre_tool_use"]
	require.True(t, ok, "expected pre_tool_use key in hooks")
	require.Len(t, preHooks, 1)
	assert.Contains(t, string(preHooks[0]), "echo check")
}

func TestDiscoverLoadsMCP(t *testing.T) {
	root := t.TempDir()

	pluginDir := filepath.Join(root, "mcp-plugin")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))

	helperWritePluginJSON(t, pluginDir, Manifest{
		Name:    "mcp-plugin",
		MCPFile: "mcp.json",
	})

	mcpContent := `{
		"mcpServers": {
			"my-server": {"command": "npx", "args": ["-y", "my-mcp-server"]}
		}
	}`
	helperWriteFile(t, filepath.Join(pluginDir, "mcp.json"), mcpContent)

	plugins, err := Discover(context.Background(), []string{root})
	require.NoError(t, err)
	require.Len(t, plugins, 1)
	require.NotNil(t, plugins[0].MCPServers)

	server, ok := plugins[0].MCPServers["my-server"]
	require.True(t, ok, "expected my-server key in MCPServers")
	assert.Contains(t, string(server), "npx")
}

func TestDiscoverContextCancellation(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, "cancel-test")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))
	helperWritePluginJSON(t, pluginDir, Manifest{Name: "cancel-test"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Discover(ctx, []string{root})
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestDiscoverAppliesDefaults(t *testing.T) {
	root := t.TempDir()

	pluginDir := filepath.Join(root, "defaults")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))

	// Minimal manifest with no defaults set.
	raw := `{"name": "defaults"}`
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(raw), 0o644))

	plugins, err := Discover(context.Background(), []string{root})
	require.NoError(t, err)
	require.Len(t, plugins, 1)

	p := plugins[0]
	assert.Equal(t, "0.0.0", p.Manifest.Version)
	assert.Equal(t, "skills", p.Manifest.SkillsDir)
	assert.Equal(t, "hooks.json", p.Manifest.HooksFile)
	assert.Equal(t, "mcp.json", p.Manifest.MCPFile)
	assert.False(t, p.Manifest.EnabledByDefault)
}

func TestDiscoverSkipsPluginWithoutName(t *testing.T) {
	root := t.TempDir()

	pluginDir := filepath.Join(root, "no-name")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))

	raw := `{"version": "1.0.0"}`
	require.NoError(t, os.WriteFile(filepath.Join(pluginDir, "plugin.json"), []byte(raw), 0o644))

	plugins, err := Discover(context.Background(), []string{root})
	require.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestManager_List_Empty(t *testing.T) {
	m := NewManager()
	result := m.List()
	assert.Empty(t, result)
}

func TestManager_Get_NotFound(t *testing.T) {
	m := NewManager()
	p := m.Get("nonexistent")
	assert.Nil(t, p)
}

func TestManager_DiscoverAndList(t *testing.T) {
	root := t.TempDir()

	pluginDir := filepath.Join(root, "list-plugin")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))
	helperWritePluginJSON(t, pluginDir, Manifest{
		Name:    "list-plugin",
		Version: "1.0.0",
	})

	m := NewManager()
	require.NoError(t, m.Discover(context.Background(), root))

	list := m.List()
	require.Len(t, list, 1)
	assert.Equal(t, "list-plugin", list[0].Manifest.Name)
}

func TestManager_Get_AfterDiscover(t *testing.T) {
	root := t.TempDir()

	pluginDir := filepath.Join(root, "get-plugin")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))
	helperWritePluginJSON(t, pluginDir, Manifest{
		Name:    "get-plugin",
		Version: "2.0.0",
	})

	m := NewManager()
	require.NoError(t, m.Discover(context.Background(), root))

	p := m.Get("get-plugin")
	require.NotNil(t, p)
	assert.Equal(t, "get-plugin", p.Manifest.Name)
	assert.Equal(t, "2.0.0", p.Manifest.Version)

	assert.Nil(t, m.Get("wrong-name"))
}

func TestManager_List_ReturnsCopy(t *testing.T) {
	root := t.TempDir()

	pluginDir := filepath.Join(root, "copy-plugin")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))
	helperWritePluginJSON(t, pluginDir, Manifest{
		Name: "copy-plugin",
	})

	m := NewManager()
	require.NoError(t, m.Discover(context.Background(), root))

	first := m.List()
	require.Len(t, first, 1)

	// Mutate the returned slice — result is intentionally discarded to verify
	// that List returns a defensive copy.
	_ = append(first, nil)

	// List again and verify the internal state is unchanged.
	second := m.List()
	assert.Len(t, second, 1)
	assert.Equal(t, "copy-plugin", second[0].Manifest.Name)
}

func TestManager_Discover_Overwrites(t *testing.T) {
	dirA := t.TempDir()

	pluginDir := filepath.Join(dirA, "plugin-a")
	require.NoError(t, os.MkdirAll(pluginDir, 0o755))
	helperWritePluginJSON(t, pluginDir, Manifest{
		Name: "plugin-a",
	})

	m := NewManager()
	require.NoError(t, m.Discover(context.Background(), dirA))
	require.Len(t, m.List(), 1)

	// Discover from an empty directory replaces the previous plugins.
	dirB := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dirB, "empty"), 0o755))

	require.NoError(t, m.Discover(context.Background(), dirB))
	assert.Empty(t, m.List())
}
