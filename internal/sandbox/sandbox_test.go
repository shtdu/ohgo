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
