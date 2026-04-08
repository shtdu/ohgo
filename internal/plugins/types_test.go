package plugins

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestJSONRoundTrip(t *testing.T) {
	original := Manifest{
		Name:             "test-plugin",
		Version:          "1.2.3",
		Description:      "A test plugin",
		EnabledByDefault: true,
		SkillsDir:        "skills",
		HooksFile:        "hooks.json",
		MCPFile:          "mcp.json",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Manifest
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.Name, decoded.Name)
	assert.Equal(t, original.Version, decoded.Version)
	assert.Equal(t, original.Description, decoded.Description)
	assert.Equal(t, original.EnabledByDefault, decoded.EnabledByDefault)
	assert.Equal(t, original.SkillsDir, decoded.SkillsDir)
	assert.Equal(t, original.HooksFile, decoded.HooksFile)
	assert.Equal(t, original.MCPFile, decoded.MCPFile)
}

func TestManifestWithMissingOptionalFields(t *testing.T) {
	raw := `{"name": "minimal"}`

	var m Manifest
	require.NoError(t, json.Unmarshal([]byte(raw), &m))

	assert.Equal(t, "minimal", m.Name)
	assert.Equal(t, "", m.Version)
	assert.Equal(t, "", m.Description)
	assert.False(t, m.EnabledByDefault)
	assert.Equal(t, "", m.SkillsDir)
	assert.Equal(t, "", m.HooksFile)
	assert.Equal(t, "", m.MCPFile)
}

func TestManifestOmitEmptyFields(t *testing.T) {
	m := Manifest{Name: "bare"}

	data, err := json.Marshal(m)
	require.NoError(t, err)

	// Version, Description, SkillsDir, HooksFile, MCPFile should be omitted
	// or zero-valued when not set. The key behavior we test is that
	// round-tripping a minimal manifest works.
	var decoded Manifest
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, m.Name, decoded.Name)
}

func TestManifestEnabledByDefaultFalse(t *testing.T) {
	raw := `{"name": "opt-out", "enabled_by_default": false}`

	var m Manifest
	require.NoError(t, json.Unmarshal([]byte(raw), &m))

	assert.Equal(t, "opt-out", m.Name)
	assert.False(t, m.EnabledByDefault)
}

func TestLoadedPluginFields(t *testing.T) {
	p := &LoadedPlugin{
		Manifest: &Manifest{
			Name:    "hello",
			Version: "0.1.0",
		},
		Path:    "/plugins/hello",
		Enabled: true,
	}

	assert.Equal(t, "hello", p.Manifest.Name)
	assert.Equal(t, "/plugins/hello", p.Path)
	assert.True(t, p.Enabled)
	assert.Nil(t, p.Skills)
	assert.Nil(t, p.Hooks)
	assert.Nil(t, p.MCPServers)
}
