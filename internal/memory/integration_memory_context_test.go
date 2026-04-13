//go:build integration

package memory_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/memory"
	"github.com/shtdu/ohgo/internal/prompts"
)

// EARS: REQ-MC-001
// Memory added in one Store instance is visible to a new Store for the same project,
// and also appears when LoadPrompt reads the MEMORY.md index.
func TestIntegration_Memory_PersistenceAndPromptGeneration(t *testing.T) {
	dir := t.TempDir()

	// Add memory through first store
	store1, err := memory.NewStore(dir)
	require.NoError(t, err)
	_, err = store1.Add("Test Note", "This is important context")
	require.NoError(t, err)

	// Second store instance for same project — cross-instance persistence
	store2, err := memory.NewStore(dir)
	require.NoError(t, err)
	files, err := store2.List()
	require.NoError(t, err)
	assert.Contains(t, files, "test_note.md")

	// Cross-component: memory.LoadPrompt produces content for prompt injection
	prompt, err := store2.LoadPrompt(100)
	require.NoError(t, err)
	assert.Contains(t, prompt, "test_note", "LoadPrompt should include memory index entries")
}

// EARS: REQ-MC-005
// Add and remove entries, verifying the MEMORY.md index stays in sync,
// and that LoadPrompt reflects the changes.
func TestIntegration_Memory_AddRemove_IndexAndPromptSync(t *testing.T) {
	dir := t.TempDir()
	store, err := memory.NewStore(dir)
	require.NoError(t, err)

	// Add two entries
	_, err = store.Add("Keep This", "persistent content")
	require.NoError(t, err)
	_, err = store.Add("Remove This", "temporary content")
	require.NoError(t, err)

	// Prompt should contain both
	prompt, err := store.LoadPrompt(100)
	require.NoError(t, err)
	assert.Contains(t, prompt, "keep_this")
	assert.Contains(t, prompt, "remove_this")

	// Remove one entry
	removed, err := store.Remove("remove_this")
	require.NoError(t, err)
	assert.True(t, removed)

	// Prompt should no longer contain removed entry
	prompt, err = store.LoadPrompt(100)
	require.NoError(t, err)
	assert.Contains(t, prompt, "keep_this")
	assert.NotContains(t, prompt, "remove_this", "removed entry should disappear from prompt")
}

// EARS: REQ-MC-002
// Memory content integrates with the prompts.Assembler via CLAUDE.md discovery.
// A CLAUDE.md file in the project dir is picked up by the prompt assembler.
func TestIntegration_Memory_CLAUDEmdDiscoveryInPrompt(t *testing.T) {
	dir := t.TempDir()

	// Write a CLAUDE.md file in the project directory
	claudeMdContent := "# Project Rules\n\nAlways use tabs for indentation."
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(claudeMdContent), 0o644))

	// Prompts assembler discovers the CLAUDE.md
	files, err := prompts.DiscoverCLAUDEmd(context.Background(), dir)
	require.NoError(t, err)
	require.NotEmpty(t, files, "CLAUDE.md should be discovered")
	assert.Contains(t, files[0].Content, "Always use tabs")

	// Merge produces a combined prompt section
	merged := prompts.MergeCLAUDEmd(files, 12000)
	require.NotNil(t, merged)
	assert.Contains(t, *merged, "Always use tabs")

	// Full assembler build includes CLAUDE.md content
	assembler := prompts.NewAssembler(dir)
	systemPrompt, err := assembler.Build(context.Background())
	require.NoError(t, err)
	assert.Contains(t, systemPrompt, "Always use tabs", "assembled prompt should include CLAUDE.md")
}

// EARS: REQ-MC-001, REQ-MC-002
// Dual-layer memory (personal + project) both appear in LoadPrompt.
func TestIntegration_Memory_DualLayerPrompt(t *testing.T) {
	dir := t.TempDir()
	store, err := memory.NewStore(dir)
	require.NoError(t, err)

	// Add to both scopes
	_, err = store.AddPersonal("Personal Note", "my personal context")
	require.NoError(t, err)
	_, err = store.Add("Project Note", "shared project context")
	require.NoError(t, err)

	prompt, err := store.LoadPrompt(200)
	require.NoError(t, err)
	assert.Contains(t, prompt, "Personal Memory", "personal section should appear")
	assert.Contains(t, prompt, "personal_note")
	assert.Contains(t, prompt, "Project Memory", "project section should appear")
	assert.Contains(t, prompt, "project_note")
}

// EARS: REQ-MC-005
// Removing nonexistent entries is a safe no-op; prompt still reflects valid entries.
func TestIntegration_Memory_RemoveNoOp_PromptUnchanged(t *testing.T) {
	dir := t.TempDir()
	store, err := memory.NewStore(dir)
	require.NoError(t, err)

	// Add one real entry
	_, err = store.Add("Real Entry", "actual content")
	require.NoError(t, err)

	// Get baseline prompt
	promptBefore, err := store.LoadPrompt(100)
	require.NoError(t, err)

	// Remove nonexistent — should be safe no-op
	removed, err := store.Remove("nonexistent")
	require.NoError(t, err)
	assert.False(t, removed)

	// Prompt unchanged
	promptAfter, err := store.LoadPrompt(100)
	require.NoError(t, err)
	assert.Equal(t, promptBefore, promptAfter)
}

// EARS: REQ-MC-002
// .claude/rules/*.md files are discovered alongside CLAUDE.md.
func TestIntegration_Memory_RulesDiscovery(t *testing.T) {
	dir := t.TempDir()

	// Create CLAUDE.md and a rules file
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Main"), 0o644))
	rulesDir := filepath.Join(dir, ".claude", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "testing.md"), []byte("# Testing Rules\nWrite tests first."), 0o644))

	files, err := prompts.DiscoverCLAUDEmd(context.Background(), dir)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 2, "should discover CLAUDE.md and rules file")

	// Merge all
	merged := prompts.MergeCLAUDEmd(files, 12000)
	require.NotNil(t, merged)
	assert.Contains(t, *merged, "Write tests first")
}
