package commands

import (
	"context"
	"os/exec"
)

// runCmd executes a command with the given arguments in cwd and returns its
// combined stdout+stderr output. It respects context cancellation.
func runCmd(ctx context.Context, name string, args []string, cwd string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	return string(out), err
}
