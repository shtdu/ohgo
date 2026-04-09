package commands

import (
	"context"
	"fmt"
	"os/exec"
)

type issueCmd struct{}

var _ Command = issueCmd{}

func (issueCmd) Name() string { return "issue" }

func (issueCmd) ShortHelp() string {
	return "List recent GitHub issues (requires gh CLI)"
}

func (issueCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return Result{}, fmt.Errorf("gh CLI not found: install from https://cli.github.com")
	}

	out, err := runCmd("gh", []string{"issue", "list", "--limit", "10"}, deps.Cwd)
	if err != nil {
		return Result{}, fmt.Errorf("gh issue list: %w", err)
	}
	if out == "" {
		return Result{Output: "No open issues."}, nil
	}
	return Result{Output: out}, nil
}
