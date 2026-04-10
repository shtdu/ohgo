package sandbox

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckAvailability(t *testing.T) {
	avail := CheckAvailability()
	// srt is installed (per README prerequisites)
	assert.True(t, avail.Available, "srt should be available")
	assert.True(t, avail.Enabled, "srt should be enabled")
	assert.NotEmpty(t, avail.Command, "should have srt path")
	assert.True(t, avail.Active(), "sandbox should be active")
}

func TestCheckAvailability_NoSrt(t *testing.T) {
	// Temporarily remove srt from PATH
	originalPath := os.Getenv("PATH")
	t.Cleanup(func() { os.Setenv("PATH", originalPath) })

	// Set PATH to empty so srt cannot be found
	t.Setenv("PATH", "/nonexistent")
	avail := CheckAvailability()
	assert.False(t, avail.Available)
	assert.False(t, avail.Enabled)
	assert.Contains(t, avail.Reason, "srt")
}

func TestAvailability_Active(t *testing.T) {
	assert.True(t, Availability{Enabled: true, Available: true}.Active())
	assert.False(t, Availability{Enabled: true, Available: false}.Active())
	assert.False(t, Availability{Enabled: false, Available: true}.Active())
}

func TestWrapCommand_Active(t *testing.T) {
	// srt is installed, so WrapCommand should wrap the command
	argv := []string{"echo", "hello"}
	wrapped, tmpFile, err := WrapCommand(argv)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	assert.Contains(t, wrapped[0], "srt", "first element should be srt binary")
	assert.NotEmpty(t, tmpFile, "should return a temp config file path")

	// Verify the wrapped command structure: srt --settings <config> -- <original args>
	require.GreaterOrEqual(t, len(wrapped), 4, "should have srt prefix args")
	assert.Equal(t, "--settings", wrapped[1])
	assert.Equal(t, tmpFile, wrapped[2])
	assert.Equal(t, "--", wrapped[3])
	assert.Equal(t, argv, wrapped[4:], "original args should be at the end")
}

func TestWrapCommand_Active_TempFileIsValid(t *testing.T) {
	argv := []string{"ls"}
	wrapped, tmpFile, err := WrapCommand(argv)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	// The temp file should contain valid JSON config
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	var parsed map[string]any
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Contains(t, parsed, "allow")
	assert.Contains(t, parsed, "deny")

	// Verify wrapped command references the same temp file
	assert.Equal(t, tmpFile, wrapped[2])
}

func TestWrapCommand_NotActive(t *testing.T) {
	// Force inactive by removing srt from PATH
	originalPath := os.Getenv("PATH")
	t.Cleanup(func() { os.Setenv("PATH", originalPath) })
	t.Setenv("PATH", "/nonexistent")

	argv := []string{"echo", "hello"}
	wrapped, tmpFile, err := WrapCommand(argv)
	require.NoError(t, err)
	assert.Equal(t, argv, wrapped, "should pass through unchanged")
	assert.Empty(t, tmpFile, "should not create temp file")
}

func TestWrapCommand_EmptyArgv(t *testing.T) {
	wrapped, tmpFile, err := WrapCommand([]string{})
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	assert.Contains(t, wrapped[0], "srt")
	assert.Equal(t, "--settings", wrapped[1])
	assert.Equal(t, "--", wrapped[3])
	// No original args appended
	assert.Equal(t, 4, len(wrapped), "should have only srt prefix args")
}

func TestWrapCommand_NilArgv(t *testing.T) {
	wrapped, tmpFile, err := WrapCommand(nil)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	assert.Contains(t, wrapped[0], "srt")
	assert.Equal(t, 4, len(wrapped), "should have only srt prefix args, nil appends nothing")
}

func TestWrapCommand_ComplexArgv(t *testing.T) {
	argv := []string{"bash", "-c", "echo 'hello world' && ls -la"}
	wrapped, tmpFile, err := WrapCommand(argv)
	require.NoError(t, err)
	defer os.Remove(tmpFile)

	// Verify original args appear at the end of the wrapped command
	tail := wrapped[len(wrapped)-len(argv):]
	assert.Equal(t, argv, tail)
}

func TestBuildConfig(t *testing.T) {
	config, err := buildConfig()
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(config, &parsed)
	require.NoError(t, err)

	allow, ok := parsed["allow"].([]any)
	require.True(t, ok)
	assert.Empty(t, allow)

	deny, ok := parsed["deny"].([]any)
	require.True(t, ok)
	assert.Empty(t, deny)
}

func TestBuildConfig_ValidJSON(t *testing.T) {
	config, err := buildConfig()
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(config, &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed, 2)
	assert.Contains(t, parsed, "allow")
	assert.Contains(t, parsed, "deny")

	allow, ok := parsed["allow"].([]any)
	assert.True(t, ok, "allow should be a JSON array")
	assert.NotNil(t, allow)

	deny, ok := parsed["deny"].([]any)
	assert.True(t, ok, "deny should be a JSON array")
	assert.NotNil(t, deny)
}

func TestAvailability_Fields(t *testing.T) {
	avail := Availability{
		Enabled:   true,
		Available: true,
		Reason:    "test",
		Command:   "/usr/bin/srt",
	}

	assert.True(t, avail.Enabled)
	assert.True(t, avail.Available)
	assert.Equal(t, "test", avail.Reason)
	assert.Equal(t, "/usr/bin/srt", avail.Command)
	assert.True(t, avail.Active())

	avail.Enabled = false
	assert.False(t, avail.Active())

	avail.Enabled = true
	avail.Available = false
	assert.False(t, avail.Active())

	zero := Availability{}
	assert.False(t, zero.Active())
	assert.Empty(t, zero.Reason)
	assert.Empty(t, zero.Command)
}
