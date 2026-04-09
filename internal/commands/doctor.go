package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type doctorCmd struct{}

var _ Command = doctorCmd{}

func (doctorCmd) Name() string { return "doctor" }

func (doctorCmd) ShortHelp() string {
	return "Show environment diagnostics (OS, Go, Git, shell, cwd)"
}

func (doctorCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	var b strings.Builder

	fmt.Fprintf(&b, "ohgo version : %s\n", deps.Version)
	fmt.Fprintf(&b, "OS            : %s\n", runtime.GOOS)
	fmt.Fprintf(&b, "Architecture  : %s\n", runtime.GOARCH)
	fmt.Fprintf(&b, "Go (runtime)  : %s\n", runtime.Version())

	if shell := os.Getenv("SHELL"); shell != "" {
		fmt.Fprintf(&b, "Shell         : %s\n", shell)
	} else {
		fmt.Fprintln(&b, "Shell         : (unknown)")
	}

	fmt.Fprintf(&b, "Working dir   : %s\n", deps.Cwd)

	// Go version from toolchain.
	if out, err := runCmd("go", []string{"version"}, deps.Cwd); err == nil {
		fmt.Fprintf(&b, "Go (toolchain): %s\n", strings.TrimSpace(out))
	} else {
		fmt.Fprintln(&b, "Go (toolchain): not found")
	}

	// Git version.
	if out, err := runCmd("git", []string{"--version"}, deps.Cwd); err == nil {
		fmt.Fprintf(&b, "Git           : %s\n", strings.TrimSpace(out))
	} else {
		fmt.Fprintln(&b, "Git           : not found")
	}

	// Check if inside a git repo.
	if _, err := runCmd("git", []string{"rev-parse", "--is-inside-work-tree"}, deps.Cwd); err == nil {
		fmt.Fprintln(&b, "Git repo      : yes")
	} else {
		fmt.Fprintln(&b, "Git repo      : no")
	}

	// Check gh CLI.
	if _, err := exec.LookPath("gh"); err == nil {
		if out, err := runCmd("gh", []string{"--version"}, deps.Cwd); err == nil {
			line := strings.SplitN(out, "\n", 2)[0]
			fmt.Fprintf(&b, "GitHub CLI    : %s\n", strings.TrimSpace(line))
		}
	} else {
		fmt.Fprintln(&b, "GitHub CLI    : not found")
	}

	return Result{Output: b.String()}, nil
}
