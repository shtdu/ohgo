package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

type filesCmd struct{}

var _ Command = filesCmd{}

func (filesCmd) Name() string { return "files" }

func (filesCmd) ShortHelp() string {
	return "List workspace files (skips hidden files and .git)"
}

func (filesCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	entries, err := os.ReadDir(deps.Cwd)
	if err != nil {
		return Result{}, fmt.Errorf("reading directory %s: %w", deps.Cwd, err)
	}

	var b strings.Builder
	tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)

	for _, e := range entries {
		name := e.Name()
		// Skip hidden files and .git directory.
		if strings.HasPrefix(name, ".") {
			continue
		}

		if e.IsDir() {
			fmt.Fprintf(tw, "%s/\t(dir)\n", name)
		} else {
			info, err := e.Info()
			if err != nil {
				fmt.Fprintf(tw, "%s\t(unknown)\n", name)
				continue
			}
			fmt.Fprintf(tw, "%s\t%d B\n", name, info.Size())
		}
	}

	if err := tw.Flush(); err != nil {
		return Result{}, fmt.Errorf("formatting output: %w", err)
	}

	out := b.String()
	if out == "" {
		out = "(no visible files)"
	}
	return Result{Output: out}, nil
}
