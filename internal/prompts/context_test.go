package prompts

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/memory"
)

func TestAssembler_FullAssembly(t *testing.T) {
	// Create a temp dir with a CLAUDE.md file.
	dir := t.TempDir()
	claudeMdContent := "# Test Project\nThis is a test project with custom instructions."
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(claudeMdContent), 0o644))

	a := NewAssembler(dir)
	result, err := a.Build(context.Background())
	require.NoError(t, err)

	// Should contain environment section from BuildSystemPrompt.
	assert.Contains(t, result, "## Environment")
	assert.Contains(t, result, "- OS:")

	// Should contain CLAUDE.md content.
	assert.Contains(t, result, claudeMdContent)
	assert.Contains(t, result, "# Project instructions from")
}

func TestAssembler_NoCLAUDEmd(t *testing.T) {
	// Empty temp dir — no CLAUDE.md files within the temp dir.
	// Note: DiscoverCLAUDEmd walks upward, so CLAUDE.md files from parent
	// directories (e.g. ~/.claude/CLAUDE.md) may still be discovered.
	// We verify only that the temp dir's own files are absent by checking
	// that the result does not reference the temp dir path in a project
	// instruction header.
	dir := t.TempDir()

	a := NewAssembler(dir)
	result, err := a.Build(context.Background())
	require.NoError(t, err)

	// Should contain the system prompt with environment section.
	assert.Contains(t, result, "## Environment")
	// No CLAUDE.md should exist inside the temp dir itself.
	assert.NotContains(t, result, "# Project instructions from "+dir)
}

func TestAssembler_CustomPromptOverride(t *testing.T) {
	dir := t.TempDir()
	// Create a CLAUDE.md so we can verify it is still appended.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("project rules"), 0o644))

	a := NewAssembler(dir).WithCustomPrompt("Custom base prompt")
	result, err := a.Build(context.Background())
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(result, "Custom base prompt"), "result should start with custom prompt")
	assert.NotContains(t, result, "You are og (OpenHarness Go)", "base prompt should not appear")
	// CLAUDE.md content should still be appended.
	assert.Contains(t, result, "project rules")
}

func TestAssembler_BuildReturnsNonEmptyString(t *testing.T) {
	dir := t.TempDir()

	a := NewAssembler(dir)
	result, err := a.Build(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestAssembler_EmptyCwdUsesCurrentDir(t *testing.T) {
	a := NewAssembler("")
	result, err := a.Build(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestAssembler_MemoryInjectedIntoPrompt(t *testing.T) {
	dir := t.TempDir()

	// Create a memory store and add a project memory entry.
	store, err := memory.NewStore(dir)
	require.NoError(t, err)
	_, err = store.Add("Test Memory", "This is remembered context")
	require.NoError(t, err)

	a := NewAssembler(dir).WithMemoryStore(store)
	result, err := a.Build(context.Background())
	require.NoError(t, err)

	assert.Contains(t, result, "Test Memory", "prompt should contain memory title from index")
}

func TestAssembler_MemoryNotInjectedWhenNil(t *testing.T) {
	dir := t.TempDir()

	// No memory store set — should not panic or add memory section headers.
	a := NewAssembler(dir)
	result, err := a.Build(context.Background())
	require.NoError(t, err)
	assert.NotContains(t, result, "Personal Memory")
	assert.NotContains(t, result, "Project Memory")
}

func TestAssembler_MemoryPersonalAndProject(t *testing.T) {
	dir := t.TempDir()

	store, err := memory.NewStore(dir)
	require.NoError(t, err)
	_, err = store.AddPersonal("My Prefs", "I prefer Go")
	require.NoError(t, err)
	_, err = store.Add("Auth Rewrite", "Compliance-driven")
	require.NoError(t, err)

	a := NewAssembler(dir).WithMemoryStore(store)
	result, err := a.Build(context.Background())
	require.NoError(t, err)

	assert.Contains(t, result, "Personal Memory", "should include personal memory section header")
	assert.Contains(t, result, "Project Memory", "should include project memory section header")
}
