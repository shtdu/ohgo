package auth

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectFlow(t *testing.T) {
	tests := []struct {
		authSource string
		wantName   string
	}{
		{"anthropic_api_key", "api_key"},
		{"openai_api_key", "api_key"},
		{"copilot_oauth", "device_code"},
		{"codex_subscription", "external_cli"},
		{"claude_subscription", "external_cli"},
		{"unknown", ""},
	}
	for _, tt := range tests {
		t.Run(tt.authSource, func(t *testing.T) {
			flow := SelectFlow(tt.authSource)
			if tt.wantName == "" {
				assert.Nil(t, flow)
			} else {
				require.NotNil(t, flow)
				assert.Equal(t, tt.wantName, flow.Name())
			}
		})
	}
}

func TestExternalCLIFlow(t *testing.T) {
	dir := t.TempDir()
	credFile := filepath.Join(dir, "credentials.json")
	credData := map[string]any{
		"token": "test-token-123",
	}
	data, err := json.Marshal(credData)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(credFile, data, 0o644))

	flow := &ExternalCLIFlow{
		Provider:  "test-provider",
		ConfigDir: dir,
		FileName:  "credentials.json",
	}

	cred, err := flow.Authenticate(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "test-provider", cred.Provider)
	assert.Equal(t, "external", cred.Kind)
	assert.Equal(t, "test-token-123", cred.Value)
}

func TestExternalCLIFlow_MissingFile(t *testing.T) {
	flow := &ExternalCLIFlow{
		Provider:  "test",
		ConfigDir: "/nonexistent/path",
		FileName:  "credentials.json",
	}

	_, err := flow.Authenticate(context.Background())
	require.Error(t, err)
}

func TestExternalCLIFlow_APIKeyField(t *testing.T) {
	dir := t.TempDir()
	credFile := filepath.Join(dir, "credentials.json")
	credData := map[string]any{
		"api_key": "sk-from-api-key-field",
	}
	data, err := json.Marshal(credData)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(credFile, data, 0o644))

	flow := &ExternalCLIFlow{
		Provider:  "test",
		ConfigDir: dir,
		FileName:  "credentials.json",
	}

	cred, err := flow.Authenticate(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "sk-from-api-key-field", cred.Value)
}

func TestCopilotDeviceFlow_Name(t *testing.T) {
	f := NewCopilotDeviceFlow()
	assert.Equal(t, "device_code", f.Name())
	assert.Equal(t, copilotClientID, f.ClientID)
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	if home != "" {
		assert.Equal(t, filepath.Join(home, "test"), expandHome("~/test"))
	}
	assert.Equal(t, "/absolute/path", expandHome("/absolute/path"))
	assert.Equal(t, "relative/path", expandHome("relative/path"))
}
