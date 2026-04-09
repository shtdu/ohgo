package commands

import (
	"context"
	"fmt"
)

type diffCmd struct{}

var _ Command = diffCmd{}

func (diffCmd) Name() string { return "diff" }

func (diffCmd) ShortHelp() string {
	return "Show git diff for the working tree"
}

func (diffCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	out, err := runCmd("git", []string{"diff"}, deps.Cwd)
	if err != nil {
		return Result{}, fmt.Errorf("git diff: %w", err)
	}
	if out == "" {
		return Result{Output: "No changes in tracked files."}, nil
	}
	return Result{Output: out}, nil
}
