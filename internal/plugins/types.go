package plugins

import (
	"encoding/json"

	"github.com/shtdu/ohgo/internal/skills"
)

// Manifest describes a plugin, loaded from plugin.json.
// Compatible with claude-code/plugins format.
type Manifest struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	Description      string `json:"description"`
	EnabledByDefault bool   `json:"enabled_by_default"`
	SkillsDir        string `json:"skills_dir"`
	HooksFile        string `json:"hooks_file"`
	MCPFile          string `json:"mcp_file"`
}

// LoadedPlugin is a fully resolved plugin with its artifacts.
type LoadedPlugin struct {
	Manifest   *Manifest
	Path       string
	Enabled    bool
	Skills     []*skills.Skill
	Hooks      map[string][]json.RawMessage // event -> raw hook definitions
	MCPServers map[string]json.RawMessage   // server name -> raw MCP config
}
