//go:build integration

package plugins_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/hooks"
	"github.com/shtdu/ohgo/internal/plugins"
	"github.com/shtdu/ohgo/internal/skills"
)

// helper: create a plugin directory with a manifest and optional artifacts.
func setupPlugin(t *testing.T, root, pluginName, manifest string, artifacts map[string]string) string {
	t.Helper()
	dir := filepath.Join(root, pluginName)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(manifest), 0o644))
	for relPath, content := range artifacts {
		fullPath := filepath.Join(dir, relPath)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0o644))
	}
	return dir
}

// EARS: REQ-EX-001, REQ-EX-002
// Plugin discovery loads skills into LoadedPlugin.Skills, ready for registration.
func TestIntegration_Plugin_DiscoveryLoadsSkills(t *testing.T) {
	dir := t.TempDir()
	skillContent := "---\nname: greet\ndescription: Greeting skill\n---\nSay hello to the user."
	setupPlugin(t, dir, "skill-plugin", `{"name": "skill-plugin", "version": "1.0.0"}`, map[string]string{
		"skills/greet.md": skillContent,
	})

	results, err := plugins.Discover(context.Background(), []string{dir})
	require.NoError(t, err)
	require.Len(t, results, 1)

	plugin := results[0]
	assert.Equal(t, "skill-plugin", plugin.Manifest.Name)
	require.Len(t, plugin.Skills, 1, "plugin should load skills from skills/ dir")
	assert.Equal(t, "greet", plugin.Skills[0].Name)
	assert.Equal(t, "plugin", plugin.Skills[0].Source)

	// Wire discovered skills into a skills.Registry — cross-component
	reg := skills.NewRegistry()
	for _, s := range plugin.Skills {
		reg.Register(s)
	}
	got := reg.Get("greet")
	require.NotNil(t, got)
	assert.Contains(t, got.Content, "Say hello")
}

// EARS: REQ-EX-005
// Plugin discovery loads hook definitions from hooks.json.
func TestIntegration_Plugin_DiscoveryLoadsHooks(t *testing.T) {
	dir := t.TempDir()
	hooksJSON := `{"pre_tool_use": [{"event": "pre_tool_use", "type": "command", "command": "echo blocked"}]}`
	setupPlugin(t, dir, "hook-plugin", `{"name": "hook-plugin", "version": "1.0.0"}`, map[string]string{
		"hooks.json": hooksJSON,
	})

	results, err := plugins.Discover(context.Background(), []string{dir})
	require.NoError(t, err)
	require.Len(t, results, 1)

	plugin := results[0]
	assert.NotNil(t, plugin.Hooks, "plugin should load hooks from hooks.json")

	// Parse raw hook definitions into hooks.Registry — cross-component
	hookDefs, ok := plugin.Hooks["pre_tool_use"]
	require.True(t, ok, "should have pre_tool_use hooks")
	require.Len(t, hookDefs, 1)

	var hookDef hooks.HookDefinition
	require.NoError(t, json.Unmarshal(hookDefs[0], &hookDef))
	assert.Equal(t, hooks.HookEventPreToolUse, hookDef.Event)
	assert.Equal(t, hooks.HookTypeCommand, hookDef.Type)
	assert.Equal(t, "echo blocked", hookDef.Command)

	// Register into hooks.Registry
	reg := hooks.NewRegistry()
	reg.Register(hookDef.Event, hookDef)
	assert.Len(t, reg.Get(hooks.HookEventPreToolUse), 1)
}

// EARS: REQ-EX-007
// Plugin discovery loads MCP server configurations from mcp.json.
func TestIntegration_Plugin_DiscoveryLoadsMCP(t *testing.T) {
	dir := t.TempDir()
	mcpJSON := `{"mcpServers": {"test-server": {"command": "npx", "args": ["-y", "some-mcp"]}}}`
	setupPlugin(t, dir, "mcp-plugin", `{"name": "mcp-plugin", "version": "1.0.0"}`, map[string]string{
		"mcp.json": mcpJSON,
	})

	results, err := plugins.Discover(context.Background(), []string{dir})
	require.NoError(t, err)
	require.Len(t, results, 1)

	plugin := results[0]
	assert.NotNil(t, plugin.MCPServers, "plugin should load MCP from mcp.json")

	serverConfig, ok := plugin.MCPServers["test-server"]
	require.True(t, ok, "should have test-server MCP config")
	assert.Contains(t, string(serverConfig), "npx")
}

// EARS: REQ-EX-001
// Multiple plugins discovered in one scan, sorted by name.
func TestIntegration_Plugin_MultiplePluginsSorted(t *testing.T) {
	dir := t.TempDir()
	setupPlugin(t, dir, "zeta", `{"name": "zeta", "version": "0.1.0"}`, nil)
	setupPlugin(t, dir, "alpha", `{"name": "alpha", "version": "0.2.0"}`, nil)

	results, err := plugins.Discover(context.Background(), []string{dir})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "alpha", results[0].Manifest.Name, "plugins should be sorted by name")
	assert.Equal(t, "zeta", results[1].Manifest.Name)
}

// EARS: REQ-EX-001
// Invalid manifests are silently skipped; valid ones still loaded.
func TestIntegration_Plugin_InvalidSkippedValidLoaded(t *testing.T) {
	dir := t.TempDir()
	// Invalid
	setupPlugin(t, dir, "bad", `{invalid`, nil)
	// Valid
	setupPlugin(t, dir, "good", `{"name": "good", "version": "1.0.0"}`, nil)

	results, err := plugins.Discover(context.Background(), []string{dir})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "good", results[0].Manifest.Name)
}

// EARS: REQ-EX-008
// Manager discovers plugins and wires their skills into a skills.Registry.
func TestIntegration_Plugin_ManagerSkillsWiredToRegistry(t *testing.T) {
	mgr := plugins.NewManager()
	dir := t.TempDir()
	setupPlugin(t, dir, "wired-plugin", `{"name": "wired-plugin", "version": "1.0.0"}`, map[string]string{
		"skills/deploy.md": "---\nname: deploy\ndescription: Deploy skill\n---\nDeploy the project.",
	})

	require.NoError(t, mgr.Discover(context.Background(), dir))

	list := mgr.List()
	require.Len(t, list, 1)

	// Wire plugin skills into a real skills.Registry
	reg := skills.NewRegistry()
	for _, p := range list {
		for _, s := range p.Skills {
			reg.Register(s)
		}
	}
	got := reg.Get("deploy")
	require.NotNil(t, got, "skill should be registered from plugin")
	assert.Equal(t, "plugin", got.Source)
}
