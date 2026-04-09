package commands

import (
	"context"
	"fmt"
	"strings"
)

type branchCmd struct{}

var _ Command = branchCmd{}

func (branchCmd) Name() string { return "branch" }

func (branchCmd) ShortHelp() string {
	return "Show current git branch and short status"
}

func (branchCmd) Run(ctx context.Context, _ string, deps *Deps) (Result, error) {
	branch, err := runCmd(ctx, "git", []string{"branch", "--show-current"}, deps.Cwd)
	if err != nil {
		return Result{}, fmt.Errorf("git branch: %w", err)
	}
	branch = strings.TrimSpace(branch)
	if branch == "" {
		// Likely detached HEAD; try rev-parse instead.
		sha, err2 := runCmd(ctx, "git", []string{"rev-parse", "--short", "HEAD"}, deps.Cwd)
		if err2 != nil {
			return Result{}, fmt.Errorf("git branch: %w", err2)
		}
		branch = "DETACHED at " + strings.TrimSpace(sha)
	}

	status, err := runCmd(ctx, "git", []string{"status", "--short"}, deps.Cwd)
	if err != nil {
		return Result{}, fmt.Errorf("git status: %w", err)
	}
	status = strings.TrimSpace(status)

	var out strings.Builder
	fmt.Fprintf(&out, "Branch: %s\n", branch)
	if status == "" {
		fmt.Fprint(&out, "Working tree clean.\n")
	} else {
		fmt.Fprintln(&out, "Changes:")
		fmt.Fprint(&out, status)
		if !strings.HasSuffix(status, "\n") {
			fmt.Fprintln(&out)
		}
	}

	return Result{Output: out.String()}, nil
}
