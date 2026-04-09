package commands

import (
	"context"
	"fmt"
)

type releaseNotesCmd struct{}

var _ Command = releaseNotesCmd{}

func (releaseNotesCmd) Name() string { return "release-notes" }

func (releaseNotesCmd) ShortHelp() string {
	return "Show release notes and changelog link"
}

func (releaseNotesCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	out := fmt.Sprintf("Release notes: See https://github.com/shtdu/ohgo/releases\n\nCurrent version: %s", deps.Version)
	return Result{Output: out}, nil
}
