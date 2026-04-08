package worktree

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

// initTestRepo creates a temporary git repo with an initial commit and returns
// its path. The caller is responsible for cleanup (t.TempDir handles this).
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "command %v failed: %s", args, string(out))
	}

	return dir
}

// --- Enter Worktree Tests ---

func TestEnterWorktreeTool_NameAndSchema(t *testing.T) {
	tool := EnterWorktreeTool{}
	assert.Equal(t, "enter_worktree", tool.Name())
	assert.Contains(t, tool.Description(), "worktree")
	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
	required := schema["required"].([]string)
	assert.Contains(t, required, "name")
}

func TestEnterWorktreeTool_CreatesWorktree(t *testing.T) {
	dir := initTestRepo(t)

	tool := EnterWorktreeTool{}
	args, _ := json.Marshal(map[string]any{
		"name":          "feature-x",
		"create_branch": true,
	})
	// Override git root detection by running from the test repo
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Created worktree")

	// Verify worktree exists
	wtPath := filepath.Join(dir, ".claude", "worktrees", "feature-x")
	info, err := os.Stat(wtPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify the branch was created
	cmd := exec.Command("git", "branch", "--list", "feature-x")
	cmd.Dir = dir
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "feature-x")
}

func TestEnterWorktreeTool_CreatesWorktreeWithoutBranch(t *testing.T) {
	dir := initTestRepo(t)

	tool := EnterWorktreeTool{}
	// Explicitly set create_branch to false
	args, _ := json.Marshal(map[string]any{
		"name":          "detached-wt",
		"create_branch": false,
	})
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	wtPath := filepath.Join(dir, ".claude", "worktrees", "detached-wt")
	_, err = os.Stat(wtPath)
	require.NoError(t, err)
}

func TestEnterWorktreeTool_MissingName(t *testing.T) {
	tool := EnterWorktreeTool{}
	args, _ := json.Marshal(map[string]any{
		"create_branch": true,
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "name is required")
}

func TestEnterWorktreeTool_CustomPath(t *testing.T) {
	dir := initTestRepo(t)
	customPath := filepath.Join(t.TempDir(), "my-worktree")

	tool := EnterWorktreeTool{}
	args, _ := json.Marshal(map[string]any{
		"name":          "custom-branch",
		"path":          customPath,
		"create_branch": true,
	})
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, customPath)

	_, err = os.Stat(customPath)
	require.NoError(t, err)
}

func TestEnterWorktreeTool_InvalidJSON(t *testing.T) {
	tool := EnterWorktreeTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

// --- Exit Worktree Tests ---

func TestExitWorktreeTool_NameAndSchema(t *testing.T) {
	tool := ExitWorktreeTool{}
	assert.Equal(t, "exit_worktree", tool.Name())
	assert.Contains(t, tool.Description(), "worktree")
	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
	required := schema["required"].([]string)
	assert.Contains(t, required, "path")
	assert.Contains(t, required, "action")
}

func TestExitWorktreeTool_RemovesWorktree(t *testing.T) {
	dir := initTestRepo(t)

	// First create a worktree
	wtPath := filepath.Join(dir, ".claude", "worktrees", "to-remove")
	cmd := exec.Command("git", "worktree", "add", "-b", "to-remove", wtPath)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "setup failed: %s", string(out))

	// Verify it exists
	_, err = os.Stat(wtPath)
	require.NoError(t, err)

	// Change to test repo so gitRoot resolves correctly
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	// Now remove it
	tool := ExitWorktreeTool{}
	args, _ := json.Marshal(map[string]any{
		"path":   wtPath,
		"action": "remove",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Removed worktree")

	// Verify it is gone
	_, err = os.Stat(wtPath)
	assert.True(t, os.IsNotExist(err))
}

func TestExitWorktreeTool_KeepAction(t *testing.T) {
	dir := initTestRepo(t)

	// Create a worktree
	wtPath := filepath.Join(dir, ".claude", "worktrees", "to-keep")
	cmd := exec.Command("git", "worktree", "add", "-b", "to-keep", wtPath)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "setup failed: %s", string(out))

	// Keep it
	tool := ExitWorktreeTool{}
	args, _ := json.Marshal(map[string]any{
		"path":   wtPath,
		"action": "keep",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "kept")

	// Verify it still exists
	_, err = os.Stat(wtPath)
	require.NoError(t, err)
}

func TestExitWorktreeTool_MissingPath(t *testing.T) {
	tool := ExitWorktreeTool{}
	args, _ := json.Marshal(map[string]any{
		"action": "remove",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "path is required")
}

func TestExitWorktreeTool_InvalidAction(t *testing.T) {
	tool := ExitWorktreeTool{}
	args, _ := json.Marshal(map[string]any{
		"path":   "/tmp/some-path",
		"action": "explode",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid action")
}

func TestExitWorktreeTool_InvalidJSON(t *testing.T) {
	tool := ExitWorktreeTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestExitWorktreeTool_ForceRemove(t *testing.T) {
	dir := initTestRepo(t)

	// Create worktree and add an uncommitted file
	wtPath := filepath.Join(dir, ".claude", "worktrees", "dirty-wt")
	cmd := exec.Command("git", "worktree", "add", "-b", "dirty-wt", wtPath)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "setup failed: %s", string(out))

	// Write an untracked file in the worktree
	err = os.WriteFile(filepath.Join(wtPath, "untracked.txt"), []byte("dirty"), 0644)
	require.NoError(t, err)

	// Change to test repo so gitRoot resolves correctly
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(origDir)

	// Force remove
	tool := ExitWorktreeTool{}
	args, _ := json.Marshal(map[string]any{
		"path":           wtPath,
		"action":         "remove",
		"discard_changes": true,
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Removed worktree")

	_, err = os.Stat(wtPath)
	assert.True(t, os.IsNotExist(err))
}

func TestEnterWorktreeTool_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tool := EnterWorktreeTool{}
	args, _ := json.Marshal(map[string]any{
		"name": "cancelled",
	})
	_, err := tool.Execute(ctx, args)
	assert.Error(t, err)
}

func TestExitWorktreeTool_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tool := ExitWorktreeTool{}
	args, _ := json.Marshal(map[string]any{
		"path":   "/tmp/some-wt",
		"action": "remove",
	})
	_, err := tool.Execute(ctx, args)
	assert.Error(t, err)
}

// Verify the tools satisfy the interface
var (
	_ tools.Tool = EnterWorktreeTool{}
	_ tools.Tool = ExitWorktreeTool{}
)
