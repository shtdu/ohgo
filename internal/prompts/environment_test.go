package prompts

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectEnvironment_HappyPath(t *testing.T) {
	ctx := context.Background()
	env, err := DetectEnvironment(ctx, "")
	require.NoError(t, err)

	assert.NotEmpty(t, env.OSName, "OSName should be set")
	assert.NotEmpty(t, env.Architecture, "Architecture should be set")
	assert.NotEmpty(t, env.Shell, "Shell should be set")
	assert.NotEmpty(t, env.WorkingDir, "WorkingDir should be set")
	assert.NotEmpty(t, env.GoVersion, "GoVersion should be set")
	assert.NotEmpty(t, env.Date, "Date should be set")
	assert.NotEmpty(t, env.HomeDir, "HomeDir should be set")

	// Date format check
	_, parseErr := time.Parse("2006-01-02", env.Date)
	assert.NoError(t, parseErr, "Date should be in YYYY-MM-DD format")
}

func TestDetectEnvironment_GitDetectionInRepo(t *testing.T) {
	// This project is a git repo, so the working directory should detect it.
	cwd, err := os.Getwd()
	require.NoError(t, err)

	ctx := context.Background()
	env, err := DetectEnvironment(ctx, cwd)
	require.NoError(t, err)

	assert.True(t, env.IsGitRepo, "should detect git repo")
	assert.NotEmpty(t, env.GitBranch, "should detect branch name")
}

func TestDetectEnvironment_GitDetectionOutsideRepo(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := context.Background()
	env, err := DetectEnvironment(ctx, tmpDir)
	require.NoError(t, err)

	assert.False(t, env.IsGitRepo, "temp dir should not be a git repo")
	assert.Empty(t, env.GitBranch, "branch should be empty outside git repo")
}

func TestDetectEnvironment_EmptyCwdUsesCurrentDir(t *testing.T) {
	ctx := context.Background()
	env, err := DetectEnvironment(ctx, "")
	require.NoError(t, err)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	assert.Equal(t, cwd, env.WorkingDir, "empty cwd should resolve to current working directory")
}

func TestDetectEnvironment_ExplicitCwd(t *testing.T) {
	tmpDir := t.TempDir()

	ctx := context.Background()
	env, err := DetectEnvironment(ctx, tmpDir)
	require.NoError(t, err)

	assert.Equal(t, tmpDir, env.WorkingDir)
}

func TestDetectEnvironment_OSNameMatchesPlatform(t *testing.T) {
	ctx := context.Background()
	env, err := DetectEnvironment(ctx, "")
	require.NoError(t, err)

	switch runtime.GOOS {
	case "darwin":
		assert.Equal(t, "macos", env.OSName)
	case "linux":
		assert.Contains(t, []string{"linux", "wsl"}, env.OSName)
	case "windows":
		assert.Equal(t, "windows", env.OSName)
	}
}

func TestDetectEnvironment_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should return quickly even with a cancelled context.
	// The function may still partially succeed for non-command operations,
	// but should not block.
	done := make(chan struct{})
	go func() {
		_, _ = DetectEnvironment(ctx, "")
		close(done)
	}()

	select {
	case <-done:
		// Function returned promptly — good.
	case <-time.After(10 * time.Second):
		t.Fatal("DetectEnvironment did not return within 10s after context cancellation")
	}
}

func TestDetectEnvironment_ShellIsBaseName(t *testing.T) {
	ctx := context.Background()
	env, err := DetectEnvironment(ctx, "")
	require.NoError(t, err)

	// Shell should not contain a slash — it's just the basename.
	assert.NotContains(t, env.Shell, "/", "Shell should be basename only")
}
