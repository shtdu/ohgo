package commands

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIfNoGit skips the test if git is not available on PATH.
func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
}

// initGitRepo creates a temporary git repo and returns its path.
func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ctx := context.Background()
	_, _ = runCmd(ctx, "git", []string{"init"}, dir)
	_, _ = runCmd(ctx, "git", []string{"config", "user.email", "test@test.com"}, dir)
	_, _ = runCmd(ctx, "git", []string{"config", "user.name", "Test"}, dir)
	return dir
}

// --- doctor command tests ---

func TestDoctor_NameAndHelp(t *testing.T) {
	var cmd doctorCmd
	assert.Equal(t, "doctor", cmd.Name())
	assert.NotEmpty(t, cmd.ShortHelp())
}

func TestDoctor_Run(t *testing.T) {
	cmd := doctorCmd{}
	deps := &Deps{
		Cwd:     t.TempDir(),
		Version: "test-v1",
	}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "test-v1")
	assert.Contains(t, res.Output, "OS")
	assert.Contains(t, res.Output, "Architecture")
	assert.Contains(t, res.Output, "Go (runtime)")
}

func TestDoctor_RunInGitRepo(t *testing.T) {
	skipIfNoGit(t)
	cmd := doctorCmd{}
	deps := &Deps{
		Cwd:     initGitRepo(t),
		Version: "test-v1",
	}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "Git repo      : yes")
}

// --- files command tests ---

func TestFiles_NameAndHelp(t *testing.T) {
	var cmd filesCmd
	assert.Equal(t, "files", cmd.Name())
	assert.NotEmpty(t, cmd.ShortHelp())
}

func TestFiles_EmptyDir(t *testing.T) {
	cmd := filesCmd{}
	deps := &Deps{Cwd: t.TempDir()}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Equal(t, "(no visible files)", res.Output)
}

func TestFiles_WithFilesAndDirs(t *testing.T) {
	dir := t.TempDir()

	// Create visible files.
	require.NoError(t, os.WriteFile(dir+"/hello.txt", []byte("hi"), 0o644))
	require.NoError(t, os.WriteFile(dir+"/readme.md", []byte("readme"), 0o644))

	// Create a visible directory.
	require.NoError(t, os.Mkdir(dir+"/subdir", 0o755))

	// Create hidden files that should be skipped.
	require.NoError(t, os.WriteFile(dir+"/.hidden", []byte("nope"), 0o644))

	// Create .git directory that should be skipped.
	require.NoError(t, os.Mkdir(dir+"/.git", 0o755))

	cmd := filesCmd{}
	deps := &Deps{Cwd: dir}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)

	assert.Contains(t, res.Output, "hello.txt")
	assert.Contains(t, res.Output, "readme.md")
	assert.Contains(t, res.Output, "subdir/")
	assert.NotContains(t, res.Output, ".hidden")
	assert.NotContains(t, res.Output, ".git")
}

func TestFiles_NonexistentDir(t *testing.T) {
	cmd := filesCmd{}
	deps := &Deps{Cwd: "/nonexistent/path/xyz123"}
	_, err := cmd.Run(context.Background(), "", deps)
	require.Error(t, err)
}

// --- diff command tests ---

func TestDiff_NameAndHelp(t *testing.T) {
	var cmd diffCmd
	assert.Equal(t, "diff", cmd.Name())
	assert.NotEmpty(t, cmd.ShortHelp())
}

func TestDiff_NotGitRepo(t *testing.T) {
	cmd := diffCmd{}
	deps := &Deps{Cwd: t.TempDir()}
	_, err := cmd.Run(context.Background(), "", deps)
	require.Error(t, err)
}

func TestDiff_CleanRepo(t *testing.T) {
	skipIfNoGit(t)
	dir := initGitRepo(t)
	// Commit an initial file so the repo is not empty.
	require.NoError(t, os.WriteFile(dir+"/README.md", []byte("# test"), 0o644))
	_, _ = runCmd(context.Background(), "git", []string{"add", "-A"}, dir)
	_, _ = runCmd(context.Background(), "git", []string{"commit", "-m", "init"}, dir)

	cmd := diffCmd{}
	deps := &Deps{Cwd: dir}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Equal(t, "No changes in tracked files.", res.Output)
}

// --- branch command tests ---

func TestBranch_NameAndHelp(t *testing.T) {
	var cmd branchCmd
	assert.Equal(t, "branch", cmd.Name())
	assert.NotEmpty(t, cmd.ShortHelp())
}

func TestBranch_NotGitRepo(t *testing.T) {
	cmd := branchCmd{}
	deps := &Deps{Cwd: t.TempDir()}
	_, err := cmd.Run(context.Background(), "", deps)
	require.Error(t, err)
}

func TestBranch_CleanRepo(t *testing.T) {
	skipIfNoGit(t)
	dir := initGitRepo(t)
	require.NoError(t, os.WriteFile(dir+"/README.md", []byte("# test"), 0o644))
	_, _ = runCmd(context.Background(), "git", []string{"add", "-A"}, dir)
	_, _ = runCmd(context.Background(), "git", []string{"commit", "-m", "init"}, dir)

	cmd := branchCmd{}
	deps := &Deps{Cwd: dir}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "Branch:")
	assert.Contains(t, res.Output, "Working tree clean")
}

// --- commit command tests ---

func TestCommit_NameAndHelp(t *testing.T) {
	var cmd commitCmd
	assert.Equal(t, "commit", cmd.Name())
	assert.NotEmpty(t, cmd.ShortHelp())
}

func TestCommit_ShowStatusNoArgs(t *testing.T) {
	skipIfNoGit(t)
	dir := initGitRepo(t)
	require.NoError(t, os.WriteFile(dir+"/README.md", []byte("# test"), 0o644))

	cmd := commitCmd{}
	deps := &Deps{Cwd: dir}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "Status:")
	assert.Contains(t, res.Output, "/commit <message>")
}

func TestCommit_WithMessage(t *testing.T) {
	skipIfNoGit(t)
	dir := initGitRepo(t)
	require.NoError(t, os.WriteFile(dir+"/hello.txt", []byte("world"), 0o644))
	// Stage the file first — /commit no longer auto-stages.
	_, _ = runCmd(context.Background(), "git", []string{"add", "-A"}, dir)

	cmd := commitCmd{}
	deps := &Deps{Cwd: dir}
	res, err := cmd.Run(context.Background(), "initial commit", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "initial commit")
}

// --- issue command tests ---

func TestIssue_NameAndHelp(t *testing.T) {
	var cmd issueCmd
	assert.Equal(t, "issue", cmd.Name())
	assert.NotEmpty(t, cmd.ShortHelp())
}

// --- pr_comments command tests ---

func TestPRComments_NameAndHelp(t *testing.T) {
	var cmd prCommentsCmd
	assert.Equal(t, "pr_comments", cmd.Name())
	assert.NotEmpty(t, cmd.ShortHelp())
}

// --- release-notes command tests ---

func TestReleaseNotes_NameAndHelp(t *testing.T) {
	var cmd releaseNotesCmd
	assert.Equal(t, "release-notes", cmd.Name())
	assert.NotEmpty(t, cmd.ShortHelp())
}

func TestReleaseNotes_Run(t *testing.T) {
	cmd := releaseNotesCmd{}
	deps := &Deps{Version: "0.1.0"}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "https://github.com/shtdu/ohgo/releases")
	assert.Contains(t, res.Output, "0.1.0")
}

// --- upgrade command tests ---

func TestUpgrade_NameAndHelp(t *testing.T) {
	var cmd upgradeCmd
	assert.Equal(t, "upgrade", cmd.Name())
	assert.NotEmpty(t, cmd.ShortHelp())
}

func TestUpgrade_Run(t *testing.T) {
	cmd := upgradeCmd{}
	deps := &Deps{Version: "0.1.0"}
	res, err := cmd.Run(context.Background(), "", deps)
	require.NoError(t, err)
	assert.Contains(t, res.Output, "go install github.com/shtdu/ohgo/cmd/og@latest")
	assert.Contains(t, res.Output, "0.1.0")
}

// --- helpers test ---

func TestRunCmd_NotFound(t *testing.T) {
	_, err := runCmd(context.Background(), "nonexistent_binary_12345", nil, t.TempDir())
	require.Error(t, err)
}

func TestRunCmd_Echo(t *testing.T) {
	skipIfNoGit(t)
	// Use a command we know exists on macOS and Linux.
	out, err := runCmd(context.Background(), "echo", []string{"hello", "world"}, t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, "hello world\n", out)
}

// --- Compile-time interface check ---

func TestGitCompileTimeChecks(t *testing.T) {
	// These lines are also done at package level via var _ = ...
	// but we verify here for test coverage.
	var _ Command = doctorCmd{}
	var _ Command = diffCmd{}
	var _ Command = branchCmd{}
	var _ Command = commitCmd{}
	var _ Command = issueCmd{}
	var _ Command = prCommentsCmd{}
	var _ Command = filesCmd{}
	var _ Command = releaseNotesCmd{}
	var _ Command = upgradeCmd{}
}

// --- Registry integration tests ---

func TestRegistry_RegisterAllCommands(t *testing.T) {
	r := NewRegistry()
	r.Register(doctorCmd{})
	r.Register(diffCmd{})
	r.Register(branchCmd{})
	r.Register(commitCmd{})
	r.Register(issueCmd{})
	r.Register(prCommentsCmd{})
	r.Register(filesCmd{})
	r.Register(releaseNotesCmd{})
	r.Register(upgradeCmd{})

	names := []string{
		"doctor", "diff", "branch", "commit",
		"issue", "pr_comments", "files",
		"release-notes", "upgrade",
	}
	for _, n := range names {
		assert.NotNil(t, r.Get(n), "command %q should be registered", n)
	}
}

func TestRegistry_LookupCommands(t *testing.T) {
	r := NewRegistry()
	r.Register(doctorCmd{})
	r.Register(commitCmd{})
	r.Register(filesCmd{})

	cmd, args, ok := r.Lookup("/doctor")
	require.True(t, ok)
	assert.Equal(t, "doctor", cmd.Name())
	assert.Equal(t, "", args)

	cmd, args, ok = r.Lookup("/commit fix the bug")
	require.True(t, ok)
	assert.Equal(t, "commit", cmd.Name())
	assert.Equal(t, "fix the bug", args)

	_, _, ok = r.Lookup("/nonexistent")
	assert.False(t, ok)
}
