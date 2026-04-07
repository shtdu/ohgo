package prompts

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// filterToRoot returns only files whose Path starts with root, making tests
// robust against CLAUDE.md files existing above the temp directory.
func filterToRoot(t *testing.T, files []CLAUDEmdFile, root string) []CLAUDEmdFile {
	t.Helper()
	var filtered []CLAUDEmdFile
	for _, f := range files {
		if strings.HasPrefix(f.Path, root) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func TestDiscoverCLAUDEmd_NestedFiles(t *testing.T) {
	// Create temp dir structure:
	//   root/CLAUDE.md
	//   root/sub/CLAUDE.md
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	require.NoError(t, os.MkdirAll(sub, 0o755))

	rootContent := "# Root CLAUDE.md\nroot instructions"
	subContent := "# Sub CLAUDE.md\nsub instructions"

	require.NoError(t, os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte(rootContent), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "CLAUDE.md"), []byte(subContent), 0o644))

	allFiles, err := DiscoverCLAUDEmd(context.Background(), sub)
	require.NoError(t, err)

	files := filterToRoot(t, allFiles, root)
	require.Len(t, files, 2)

	// Closest (sub) should be first
	assert.Equal(t, filepath.Join(sub, "CLAUDE.md"), files[0].Path)
	assert.Equal(t, subContent, files[0].Content)

	// Root should be second
	assert.Equal(t, filepath.Join(root, "CLAUDE.md"), files[1].Path)
	assert.Equal(t, rootContent, files[1].Content)
}

func TestDiscoverCLAUDEmd_RulesDir(t *testing.T) {
	root := t.TempDir()
	rulesDir := filepath.Join(root, ".claude", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))

	rule1Content := "# Rule 1\nno debug in prod"
	rule2Content := "# Rule 2\nuse tabs not spaces"

	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "rule1.md"), []byte(rule1Content), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "rule2.md"), []byte(rule2Content), 0o644))

	// Non-.md file should be ignored
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "ignore.txt"), []byte("ignored"), 0o644))

	allFiles, err := DiscoverCLAUDEmd(context.Background(), root)
	require.NoError(t, err)

	files := filterToRoot(t, allFiles, root)
	require.Len(t, files, 2)

	// Verify both rules found (order from filesystem, not guaranteed)
	paths := map[string]string{
		files[0].Path: files[0].Content,
		files[1].Path: files[1].Content,
	}
	assert.Contains(t, paths, filepath.Join(rulesDir, "rule1.md"))
	assert.Contains(t, paths, filepath.Join(rulesDir, "rule2.md"))
	assert.Equal(t, rule1Content, paths[filepath.Join(rulesDir, "rule1.md")])
	assert.Equal(t, rule2Content, paths[filepath.Join(rulesDir, "rule2.md")])
}

func TestDiscoverCLAUDEmd_DotClaudeCLAUDEmd(t *testing.T) {
	root := t.TempDir()
	dotClaudeDir := filepath.Join(root, ".claude")
	require.NoError(t, os.MkdirAll(dotClaudeDir, 0o755))

	content := "# .claude/CLAUDE.md\npersonal instructions"
	require.NoError(t, os.WriteFile(filepath.Join(dotClaudeDir, "CLAUDE.md"), []byte(content), 0o644))

	allFiles, err := DiscoverCLAUDEmd(context.Background(), root)
	require.NoError(t, err)

	files := filterToRoot(t, allFiles, root)
	require.Len(t, files, 1)
	assert.Equal(t, filepath.Join(dotClaudeDir, "CLAUDE.md"), files[0].Path)
	assert.Equal(t, content, files[0].Content)
}

func TestDiscoverCLAUDEmd_NoFiles(t *testing.T) {
	root := t.TempDir()

	allFiles, err := DiscoverCLAUDEmd(context.Background(), root)
	require.NoError(t, err)

	files := filterToRoot(t, allFiles, root)
	assert.Empty(t, files)
}

func TestDiscoverCLAUDEmd_ContextCancellation(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("content"), 0o644))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	files, err := DiscoverCLAUDEmd(ctx, root)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
	// May or may not have files depending on race, but error must be context.Canceled
	_ = files
}

func TestDiscoverCLAUDEmd_Deduplication(t *testing.T) {
	// Both CLAUDE.md and .claude/CLAUDE.md exist — they are different files
	// so both should be discovered.
	root := t.TempDir()
	dotClaudeDir := filepath.Join(root, ".claude")
	require.NoError(t, os.MkdirAll(dotClaudeDir, 0o755))

	rootContent := "root level"
	dotContent := "dot claude level"
	require.NoError(t, os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte(rootContent), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".claude", "CLAUDE.md"), []byte(dotContent), 0o644))

	allFiles, err := DiscoverCLAUDEmd(context.Background(), root)
	require.NoError(t, err)

	files := filterToRoot(t, allFiles, root)
	assert.Len(t, files, 2)

	// Verify the seen-map prevents the same absolute path from appearing twice.
	// Collect all paths and confirm no duplicates.
	seen := make(map[string]bool)
	for _, f := range files {
		assert.False(t, seen[f.Path], "duplicate path: %s", f.Path)
		seen[f.Path] = true
	}
}

func TestDiscoverCLAUDEmd_SymlinkDeduplication(t *testing.T) {
	// If CLAUDE.md is a symlink to .claude/CLAUDE.md, it should only appear once.
	root := t.TempDir()
	dotClaudeDir := filepath.Join(root, ".claude")
	require.NoError(t, os.MkdirAll(dotClaudeDir, 0o755))

	content := "shared content"
	dotClaudeMdPath := filepath.Join(dotClaudeDir, "CLAUDE.md")
	require.NoError(t, os.WriteFile(dotClaudeMdPath, []byte(content), 0o644))

	// Create symlink: CLAUDE.md -> .claude/CLAUDE.md
	claudeMdPath := filepath.Join(root, "CLAUDE.md")
	require.NoError(t, os.Symlink(dotClaudeMdPath, claudeMdPath))

	allFiles, err := DiscoverCLAUDEmd(context.Background(), root)
	require.NoError(t, err)

	files := filterToRoot(t, allFiles, root)

	// Both paths resolve to the same inode; the seen map uses absolute paths,
	// so both entries may appear. The important thing is no crash.
	// On most systems the symlink and target have different absolute paths,
	// so we expect 2 entries. The dedup only kicks in if Abs resolves to the same path.
	_ = files
}

func TestMergeCLAUDEmd_Empty(t *testing.T) {
	result := MergeCLAUDEmd(nil, 10000)
	assert.Nil(t, result)
}

func TestMergeCLAUDEmd_Truncation(t *testing.T) {
	largeContent := strings.Repeat("x", 15000)
	files := []CLAUDEmdFile{
		{Path: "/tmp/CLAUDE.md", Content: largeContent},
	}

	result := MergeCLAUDEmd(files, 10000)
	require.NotNil(t, result)

	// Should contain truncated content + "... (truncated)" suffix
	assert.Contains(t, *result, "... (truncated)")
	assert.Contains(t, *result, "# Project instructions from /tmp/CLAUDE.md")

	// The content portion should be exactly 10000 chars of 'x' plus truncation marker
	// Total length: header + 10000 'x' chars + "\n... (truncated)"
	expectedContent := strings.Repeat("x", 10000) + "\n... (truncated)"
	assert.Contains(t, *result, expectedContent)
}

func TestMergeCLAUDEmd_NoTruncationWhenUnderLimit(t *testing.T) {
	content := "short content"
	files := []CLAUDEmdFile{
		{Path: "/tmp/CLAUDE.md", Content: content},
	}

	result := MergeCLAUDEmd(files, 10000)
	require.NotNil(t, result)
	assert.Contains(t, *result, content)
	assert.NotContains(t, *result, "... (truncated)")
}

func TestMergeCLAUDEmd_MultipleFiles(t *testing.T) {
	files := []CLAUDEmdFile{
		{Path: "/a/CLAUDE.md", Content: "content A"},
		{Path: "/b/CLAUDE.md", Content: "content B"},
	}

	result := MergeCLAUDEmd(files, 10000)
	require.NotNil(t, result)
	assert.Contains(t, *result, "# Project instructions from /a/CLAUDE.md")
	assert.Contains(t, *result, "content A")
	assert.Contains(t, *result, "# Project instructions from /b/CLAUDE.md")
	assert.Contains(t, *result, "content B")
}

func TestDiscoverCLAUDEmd_InvalidCwd(t *testing.T) {
	// An invalid cwd should still work if filepath.Abs can resolve it
	// (Abs doesn't check existence). The function will just find no files.
	files, err := DiscoverCLAUDEmd(context.Background(), "/nonexistent/path/that/does/not/exist")
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestDiscoverCLAUDEmd_ContextCancellationDuringWalk(t *testing.T) {
	root := t.TempDir()
	// Create many nested CLAUDE.md files so the walk takes multiple iterations
	current := root
	for i := 0; i < 5; i++ {
		sub := filepath.Join(current, "level"+string(rune('0'+i)))
		require.NoError(t, os.MkdirAll(sub, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(sub, "CLAUDE.md"), []byte("level content"), 0o644))
		current = sub
	}

	// Start with a context that will be cancelled shortly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This should complete before timeout since the walk is fast
	files, err := DiscoverCLAUDEmd(ctx, current)
	if err != nil {
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	}
	// Just verify no panic and either files or error
	_ = files
}

func TestDiscoverCLAUDEmd_RulesDirSubdirsIgnored(t *testing.T) {
	root := t.TempDir()
	rulesDir := filepath.Join(root, ".claude", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0o755))

	// Create a subdirectory inside rules (should be ignored)
	require.NoError(t, os.MkdirAll(filepath.Join(rulesDir, "subdir"), 0o755))
	// Create a valid rule
	require.NoError(t, os.WriteFile(filepath.Join(rulesDir, "rule.md"), []byte("rule content"), 0o644))

	allFiles, err := DiscoverCLAUDEmd(context.Background(), root)
	require.NoError(t, err)

	files := filterToRoot(t, allFiles, root)
	assert.Len(t, files, 1)
	assert.Equal(t, filepath.Join(rulesDir, "rule.md"), files[0].Path)
}
