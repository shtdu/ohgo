package config

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/tools"
)

func testSettings() *config.Settings {
	s := config.DefaultSettings()
	return &s
}

func TestConfigTool_Name(t *testing.T) {
	tool := ConfigTool{}
	assert.Equal(t, "config", tool.Name())
}

func TestConfigTool_Description(t *testing.T) {
	tool := ConfigTool{}
	assert.NotEmpty(t, tool.Description())
}

func TestConfigTool_InputSchema(t *testing.T) {
	tool := ConfigTool{}
	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])

	props := schema["properties"].(map[string]any)
	assert.Contains(t, props, "action")
	assert.Contains(t, props, "key")

	action := props["action"].(map[string]any)
	assert.Contains(t, action, "enum")

	required := schema["required"].([]string)
	assert.Contains(t, required, "action")
}

func TestConfigTool_ShowAllSettings(t *testing.T) {
	tool := ConfigTool{Settings: testSettings()}
	args, _ := json.Marshal(map[string]string{"action": "show"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	// Should be valid JSON containing expected fields.
	var parsed map[string]any
	require.NoError(t, json.Unmarshal([]byte(result.Content), &parsed))
	assert.Contains(t, parsed, "model")
	assert.Contains(t, parsed, "max_tokens")
}

func TestConfigTool_ShowSpecificKey(t *testing.T) {
	tool := ConfigTool{Settings: testSettings()}
	args, _ := json.Marshal(map[string]string{"action": "show", "key": "model"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "claude-sonnet-4-6")
}

func TestConfigTool_ShowSpecificKeyMaxTokens(t *testing.T) {
	tool := ConfigTool{Settings: testSettings()}
	args, _ := json.Marshal(map[string]string{"action": "show", "key": "max_tokens"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "16384")
}

func TestConfigTool_NilSettings(t *testing.T) {
	tool := ConfigTool{Settings: nil}
	args, _ := json.Marshal(map[string]string{"action": "show"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not available")
}

func TestConfigTool_InvalidJSON(t *testing.T) {
	tool := ConfigTool{Settings: testSettings()}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestConfigTool_UnknownAction(t *testing.T) {
	tool := ConfigTool{Settings: testSettings()}
	args, _ := json.Marshal(map[string]string{"action": "delete"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "unknown action")
}

func TestConfigTool_UnknownKey(t *testing.T) {
	tool := ConfigTool{Settings: testSettings()}
	args, _ := json.Marshal(map[string]string{"action": "show", "key": "nonexistent_field"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "unknown setting key")
}

// Verify the tool satisfies the interface.
var _ tools.Tool = ConfigTool{}
