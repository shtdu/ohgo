package commands

import (
	"context"
	"fmt"
	"os/exec"
)

type prCommentsCmd struct{}

var _ Command = prCommentsCmd{}

func (prCommentsCmd) Name() string { return "pr_comments" }

func (prCommentsCmd) ShortHelp() string {
	return "List recent GitHub pull requests (requires gh CLI)"
}

func (prCommentsCmd) Run(ctx context.Context, _ string, deps *Deps) (Result, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return Result{}, fmt.Errorf("gh CLI not found: install from https://cli.github.com")
	}

	out, err := runCmd(ctx, "gh", []string{"pr", "list", "--limit", "5"}, deps.Cwd)
	if err != nil {
		return Result{}, fmt.Errorf("gh pr list: %w", err)
	}
	if out == "" {
		return Result{Output: "No open pull requests."}, nil
	}
	return Result{Output: out}, nil
}
