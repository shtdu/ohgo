package commands

import (
	"context"
	"fmt"
	"strings"
)

type commitCmd struct{}

var _ Command = commitCmd{}

func (commitCmd) Name() string { return "commit" }

func (commitCmd) ShortHelp() string {
	return "Git commit workflow: no args shows status+diff; with args commits all changes"
}

func (commitCmd) Run(ctx context.Context, args string, deps *Deps) (Result, error) {
	msg := strings.TrimSpace(args)

	if msg == "" {
		// Show status and diff so the user can review before committing.
		status, err := runCmd(ctx, "git", []string{"status", "--short"}, deps.Cwd)
		if err != nil {
			return Result{}, fmt.Errorf("git status: %w", err)
		}
		diff, err := runCmd(ctx, "git", []string{"diff"}, deps.Cwd)
		if err != nil {
			return Result{}, fmt.Errorf("git diff: %w", err)
		}
		diffStaged, err := runCmd(ctx, "git", []string{"diff", "--cached"}, deps.Cwd)
		if err != nil {
			return Result{}, fmt.Errorf("git diff --cached: %w", err)
		}

		var b strings.Builder
		fmt.Fprintln(&b, "Status:")
		if strings.TrimSpace(status) == "" {
			fmt.Fprintln(&b, "  (clean)")
		} else {
			fmt.Fprint(&b, status)
		}
		if strings.TrimSpace(diff) != "" {
			fmt.Fprintln(&b, "\nUnstaged changes:")
			fmt.Fprint(&b, diff)
		}
		if strings.TrimSpace(diffStaged) != "" {
			fmt.Fprintln(&b, "\nStaged changes:")
			fmt.Fprint(&b, diffStaged)
		}
		fmt.Fprintln(&b, "\nProvide a commit message: /commit <message>")
		return Result{Output: b.String()}, nil
	}

	// Commit only already-staged changes. User should stage files explicitly.
	out, err := runCmd(ctx, "git", []string{"commit", "-m", msg}, deps.Cwd)
	if err != nil {
		return Result{}, fmt.Errorf("git commit: %s%s", out, err)
	}
	return Result{Output: strings.TrimSpace(out)}, nil
}
