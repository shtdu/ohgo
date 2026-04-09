package commands

import "os/exec"

// runCmd executes a command with the given arguments in cwd and returns its
// combined stdout+stderr output. It is used by the git/gh slash commands.
func runCmd(name string, args []string, cwd string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = cwd
	out, err := cmd.CombinedOutput()
	return string(out), err
}
