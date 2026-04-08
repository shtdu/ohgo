package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shtdu/ohgo/internal/skills"
)

// Discover scans directories for plugin.json manifests and loads each plugin.
// Errors encountered while loading individual plugins are logged and skipped;
// the function returns all successfully loaded plugins sorted by name.
func Discover(ctx context.Context, dirs []string) ([]*LoadedPlugin, error) {
	var plugins []*LoadedPlugin

	for _, dir := range dirs {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("discover plugins: %w", ctx.Err())
		default:
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read plugin directory %s: %w", dir, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			pluginDir := filepath.Join(dir, entry.Name())
			manifestPath := findManifest(pluginDir)
			if manifestPath == "" {
				continue
			}

			plugin, err := loadPlugin(ctx, pluginDir, manifestPath)
			if err != nil {
				// Skip plugins that fail to load.
				continue
			}

			plugins = append(plugins, plugin)
		}
	}

	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Manifest.Name < plugins[j].Manifest.Name
	})

	return plugins, nil
}

// findManifest looks for plugin.json in standard locations within a directory.
// Checks pluginDir/plugin.json and pluginDir/.claude-plugin/plugin.json.
func findManifest(pluginDir string) string {
	candidates := []string{
		filepath.Join(pluginDir, "plugin.json"),
		filepath.Join(pluginDir, ".claude-plugin", "plugin.json"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

// loadPlugin reads a single plugin directory using the given manifest path.
func loadPlugin(ctx context.Context, pluginDir string, manifestPath string) (*LoadedPlugin, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", manifestPath, err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest %s: %w", manifestPath, err)
	}

	if strings.TrimSpace(manifest.Name) == "" {
		return nil, fmt.Errorf("manifest %s: name is required", manifestPath)
	}

	// Apply defaults.
	if manifest.Version == "" {
		manifest.Version = "0.0.0"
	}
	if manifest.SkillsDir == "" {
		manifest.SkillsDir = "skills"
	}
	if manifest.HooksFile == "" {
		manifest.HooksFile = "hooks.json"
	}
	if manifest.MCPFile == "" {
		manifest.MCPFile = "mcp.json"
	}

	// Load artifacts.
	pluginSkills, err := loadPluginSkills(ctx, filepath.Join(pluginDir, manifest.SkillsDir))
	if err != nil {
		return nil, fmt.Errorf("load skills for plugin %s: %w", manifest.Name, err)
	}

	hooks, err := loadPluginHooks(filepath.Join(pluginDir, manifest.HooksFile))
	if err != nil {
		return nil, fmt.Errorf("load hooks for plugin %s: %w", manifest.Name, err)
	}

	mcp, err := loadPluginMCP(filepath.Join(pluginDir, manifest.MCPFile))
	if err != nil {
		return nil, fmt.Errorf("load MCP for plugin %s: %w", manifest.Name, err)
	}

	// Mark skills as plugin-sourced.
	for _, s := range pluginSkills {
		s.Source = "plugin"
	}

	return &LoadedPlugin{
		Manifest:   &manifest,
		Path:       pluginDir,
		Enabled:    manifest.EnabledByDefault,
		Skills:     pluginSkills,
		Hooks:      hooks,
		MCPServers: mcp,
	}, nil
}

// loadPluginSkills loads *.md files from a directory as skills.
// Returns an empty slice if the directory does not exist.
func loadPluginSkills(ctx context.Context, dir string) ([]*skills.Skill, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}
	loader := skills.NewLoader(dir)
	return loader.LoadAll(ctx)
}

// loadPluginHooks reads hooks from a JSON file.
// The file is expected to contain a map of event names to arrays of hook definitions.
// Returns raw JSON messages for later parsing by the hooks package.
func loadPluginHooks(path string) (map[string][]json.RawMessage, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read hooks file %s: %w", path, err)
	}

	// Try structured format first: {"hooks": {"event": [...]}}
	var structured struct {
		Hooks map[string][]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(data, &structured); err == nil && len(structured.Hooks) > 0 {
		return structured.Hooks, nil
	}

	// Try flat format: {"event": [...], ...}
	var flat map[string][]json.RawMessage
	if err := json.Unmarshal(data, &flat); err != nil {
		return nil, fmt.Errorf("parse hooks file %s: %w", path, err)
	}

	return flat, nil
}

// loadPluginMCP reads MCP server config from a JSON file.
// Extracts the "mcpServers" key if present, otherwise treats the whole
// object as a server map.
func loadPluginMCP(path string) (map[string]json.RawMessage, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read MCP file %s: %w", path, err)
	}

	// Try to extract "mcpServers" key.
	var wrapper struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && len(wrapper.MCPServers) > 0 {
		return wrapper.MCPServers, nil
	}

	// Fall back: treat entire object as server map.
	var servers map[string]json.RawMessage
	if err := json.Unmarshal(data, &servers); err != nil {
		return nil, fmt.Errorf("parse MCP file %s: %w", path, err)
	}

	return servers, nil
}
