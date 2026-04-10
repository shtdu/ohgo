package sandbox

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckAvailability(t *testing.T) {
	avail := CheckAvailability()
	// Result depends on whether srt is installed
	if _, err := exec.LookPath("srt"); err == nil {
		assert.True(t, avail.Available)
		assert.NotEmpty(t, avail.Command)
	} else {
		assert.False(t, avail.Available)
		assert.Contains(t, avail.Reason, "srt")
	}
}

func TestAvailability_Active(t *testing.T) {
	assert.True(t, Availability{Enabled: true, Available: true}.Active())
	assert.False(t, Availability{Enabled: true, Available: false}.Active())
	assert.False(t, Availability{Enabled: false, Available: true}.Active())
}

func TestWrapCommand_NotActive(t *testing.T) {
	// If srt is not installed, command should pass through unchanged
	argv := []string{"echo", "hello"}
	wrapped, tmpFile, err := WrapCommand(argv)
	require.NoError(t, err)

	if _, lookErr := exec.LookPath("srt"); lookErr != nil {
		// srt not installed: passthrough
		assert.Equal(t, argv, wrapped)
		assert.Empty(t, tmpFile)
	} else {
		// srt installed: wrapped
		assert.Contains(t, wrapped[0], "srt")
		assert.NotEmpty(t, tmpFile)
		_ = os.Remove(tmpFile)
	}
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

	// Verify it is valid JSON that can be re-unmarshaled
	var parsed map[string]any
	err = json.Unmarshal(config, &parsed)
	require.NoError(t, err)

	// Verify exactly the expected top-level keys exist
	assert.Len(t, parsed, 2)
	assert.Contains(t, parsed, "allow")
	assert.Contains(t, parsed, "deny")

	// Verify both allow and deny are arrays (not nil or other types)
	allow, ok := parsed["allow"].([]any)
	assert.True(t, ok, "allow should be a JSON array")
	assert.NotNil(t, allow)

	deny, ok := parsed["deny"].([]any)
	assert.True(t, ok, "deny should be a JSON array")
	assert.NotNil(t, deny)
}

func TestWrapCommand_EmptyArgv(t *testing.T) {
	// Empty argv should pass through without crashing
	wrapped, tmpFile, err := WrapCommand([]string{})
	require.NoError(t, err)

	if _, lookErr := exec.LookPath("srt"); lookErr != nil {
		// srt not installed: passthrough
		assert.Equal(t, []string{}, wrapped)
		assert.Empty(t, tmpFile)
	} else {
		// srt installed: wrapped with prefix args
		assert.True(t, len(wrapped) >= 4, "wrapped command should have srt prefix args")
		assert.NotEmpty(t, tmpFile)
		_ = os.Remove(tmpFile)
	}
}

func TestWrapCommand_NilArgv(t *testing.T) {
	// Nil argv should pass through without crashing
	wrapped, tmpFile, err := WrapCommand(nil)
	require.NoError(t, err)

	if _, lookErr := exec.LookPath("srt"); lookErr != nil {
		// srt not installed: passthrough returns nil
		assert.Nil(t, wrapped)
		assert.Empty(t, tmpFile)
	} else {
		// srt installed: wrapped with prefix args, nil gets appended as nothing
		assert.True(t, len(wrapped) >= 4, "wrapped command should have srt prefix args")
		assert.NotEmpty(t, tmpFile)
		_ = os.Remove(tmpFile)
	}
}

func TestWrapCommand_ComplexArgv(t *testing.T) {
	argv := []string{"bash", "-c", "echo 'hello world' && ls -la"}
	wrapped, tmpFile, err := WrapCommand(argv)
	require.NoError(t, err)

	if _, lookErr := exec.LookPath("srt"); lookErr != nil {
		// srt not installed: passthrough unchanged
		assert.Equal(t, argv, wrapped)
		assert.Empty(t, tmpFile)
	} else {
		// srt installed: result has at least the original args plus srt prefix
		assert.GreaterOrEqual(t, len(wrapped), len(argv))
		assert.Contains(t, wrapped[0], "srt")
		assert.NotEmpty(t, tmpFile)
		// Verify original args appear at the end of the wrapped command
		tail := wrapped[len(wrapped)-len(argv):]
		assert.Equal(t, argv, tail)
		_ = os.Remove(tmpFile)
	}
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

	// Verify toggling Enabled flips Active
	avail.Enabled = false
	assert.False(t, avail.Active())

	// Verify toggling Available flips Active
	avail.Enabled = true
	avail.Available = false
	assert.False(t, avail.Active())

	// Verify zero-value struct
	zero := Availability{}
	assert.False(t, zero.Active())
	assert.Empty(t, zero.Reason)
	assert.Empty(t, zero.Command)
}
