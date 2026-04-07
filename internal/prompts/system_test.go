package prompts

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSystemPrompt_DefaultContainsKeySections(t *testing.T) {
	env := &EnvironmentInfo{
		OSName:       "darwin",
		OSVersion:    "25.3.0",
		Architecture: "arm64",
		Shell:        "zsh",
		WorkingDir:   "/home/user/project",
		Date:         "2026-04-07",
		GoVersion:    "go1.25.6",
		IsGitRepo:    false,
	}

	result, err := BuildSystemPrompt(context.Background(), "", env)
	require.NoError(t, err)

	assert.Contains(t, result, "og")
	assert.Contains(t, result, "## Environment")
	assert.Contains(t, result, "- OS:")
	assert.Contains(t, result, "- Shell:")
	assert.Contains(t, result, "Core Capabilities")
	assert.Contains(t, result, "Key Principles")
	assert.Contains(t, result, "Tool Usage")
	assert.Contains(t, result, "Safety")
}

func TestBuildSystemPrompt_CustomPromptReplacesBase(t *testing.T) {
	env := &EnvironmentInfo{
		OSName:       "linux",
		Architecture: "amd64",
		Shell:        "bash",
		WorkingDir:   "/tmp",
		Date:         "2026-04-07",
		GoVersion:    "go1.25.6",
		IsGitRepo:    false,
	}

	result, err := BuildSystemPrompt(context.Background(), "Custom instructions", env)
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(result, "Custom instructions"), "result should start with custom prompt")
	assert.NotContains(t, result, "You are og (OpenHarness Go)", "base prompt should not appear")
}

func TestBuildSystemPrompt_PreBuiltEnvUsed(t *testing.T) {
	env := &EnvironmentInfo{
		OSName:       "darwin",
		OSVersion:    "99.0.0",
		Architecture: "arm64",
		Shell:        "fish",
		WorkingDir:   "/custom/path",
		Date:         "2026-12-25",
		GoVersion:    "go1.99.0",
		IsGitRepo:    false,
	}

	result, err := BuildSystemPrompt(context.Background(), "", env)
	require.NoError(t, err)

	assert.Contains(t, result, "darwin 99.0.0")
	assert.Contains(t, result, "arm64")
	assert.Contains(t, result, "fish")
	assert.Contains(t, result, "/custom/path")
	assert.Contains(t, result, "2026-12-25")
	assert.Contains(t, result, "go1.99.0")
}

func TestBuildSystemPrompt_EnvironmentSectionPopulated(t *testing.T) {
	env := &EnvironmentInfo{
		OSName:       "linux",
		Architecture: "amd64",
		Shell:        "bash",
		WorkingDir:   "/home/test",
		Date:         "2026-01-01",
		GoVersion:    "go1.25.0",
		IsGitRepo:    false,
	}

	result, err := BuildSystemPrompt(context.Background(), "", env)
	require.NoError(t, err)

	assert.Contains(t, result, "- OS: linux")
	assert.Contains(t, result, "- Architecture: amd64")
	assert.Contains(t, result, "- Shell: bash")
	assert.Contains(t, result, "- Working directory: /home/test")
	assert.Contains(t, result, "- Date: 2026-01-01")
	assert.Contains(t, result, "- Go version: go1.25.0")
}

func TestBuildSystemPrompt_GitBranchWhenInRepo(t *testing.T) {
	env := &EnvironmentInfo{
		OSName:       "darwin",
		Architecture: "arm64",
		Shell:        "zsh",
		WorkingDir:   "/project",
		Date:         "2026-04-07",
		GoVersion:    "go1.25.6",
		IsGitRepo:    true,
		GitBranch:    "feature/test",
	}

	result, err := BuildSystemPrompt(context.Background(), "", env)
	require.NoError(t, err)

	assert.Contains(t, result, "- Git branch: feature/test")
}

func TestBuildSystemPrompt_NoGitBranchWhenNotInRepo(t *testing.T) {
	env := &EnvironmentInfo{
		OSName:       "darwin",
		Architecture: "arm64",
		Shell:        "zsh",
		WorkingDir:   "/tmp",
		Date:         "2026-04-07",
		GoVersion:    "go1.25.6",
		IsGitRepo:    false,
		GitBranch:    "",
	}

	result, err := BuildSystemPrompt(context.Background(), "", env)
	require.NoError(t, err)

	assert.NotContains(t, result, "Git branch:")
}

func TestBuildSystemPrompt_NilEnvAutoDetects(t *testing.T) {
	// Passing nil env triggers auto-detection; this test just verifies
	// it does not error in a real environment.
	result, err := BuildSystemPrompt(context.Background(), "", nil)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	assert.Contains(t, result, "## Environment")
	assert.Contains(t, result, "- OS:")
	assert.Contains(t, result, "- Shell:")
}
