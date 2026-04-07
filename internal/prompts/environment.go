package prompts

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shtdu/ohgo/internal/config"
)

// EnvironmentInfo is a snapshot of the runtime environment.
type EnvironmentInfo struct {
	OSName       string
	OSVersion    string
	Architecture string
	Shell        string
	WorkingDir   string
	HomeDir      string
	Date         string
	GoVersion    string
	IsGitRepo    bool
	GitBranch    string
}

// DetectEnvironment gathers environment information.
// If cwd is empty, uses the current working directory.
func DetectEnvironment(ctx context.Context, cwd string) (EnvironmentInfo, error) {
	env := EnvironmentInfo{
		OSName:       string(config.DetectPlatform()),
		Architecture: runtime.GOARCH,
		Shell:        filepath.Base(config.Shell()),
		GoVersion:    runtime.Version(),
		Date:         time.Now().UTC().Format("2006-01-02"),
	}

	// Working directory
	if cwd == "" {
		dir, err := config.WorkingDir()
		if err != nil {
			return env, fmt.Errorf("detect working dir: %w", err)
		}
		env.WorkingDir = dir
	} else {
		env.WorkingDir = cwd
	}

	// Home directory
	if home, err := os.UserHomeDir(); err == nil {
		env.HomeDir = home
	}

	// OS version
	env.OSVersion = detectOSVersion(ctx)

	// Git info
	env.IsGitRepo, env.GitBranch = detectGitInfo(ctx, env.WorkingDir)

	return env, nil
}

// detectOSVersion returns the OS version string.
func detectOSVersion(ctx context.Context) string {
	switch runtime.GOOS {
	case "darwin":
		return detectMacOSVersion(ctx)
	case "linux":
		return detectLinuxVersion()
	default:
		return ""
	}
}

func detectMacOSVersion(ctx context.Context) string {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "sw_vers", "-productVersion").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func detectLinuxVersion() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if v, ok := strings.CutPrefix(line, "VERSION="); ok {
			return strings.Trim(v, `"`)
		}
	}
	return ""
}

// detectGitInfo checks if cwd is inside a git repo and returns the current branch.
func detectGitInfo(ctx context.Context, cwd string) (isRepo bool, branch string) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var buf bytes.Buffer
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = cwd
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return false, ""
	}
	if strings.TrimSpace(buf.String()) != "true" {
		return false, ""
	}

	buf.Reset()
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = cwd
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return true, ""
	}
	return true, strings.TrimSpace(buf.String())
}
