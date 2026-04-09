package ui

import (
	"bufio"
	"context"
	"strings"
)

// PermissionPrompter displays permission requests to the user.
type PermissionPrompter struct {
	in  *bufio.Reader
	out ioWriter
}

type ioWriter = interface {
	Printf(format string, args ...any)
}

// NewPermissionPrompter creates a prompter reading from the given input.
func NewPermissionPrompter(in *bufio.Reader, out ioWriter) *PermissionPrompter {
	return &PermissionPrompter{in: in, out: out}
}

// PromptApproval asks the user to approve or deny a tool execution.
// Returns (allow, remember, error).
// User input: y/yes = allow, n/no = deny, always = allow + remember.
func (p *PermissionPrompter) PromptApproval(ctx context.Context, toolName string, details string) (bool, bool, error) {
	p.out.Printf("Permission required: tool %q wants to execute.\n", toolName)
	if details != "" && len(details) <= 200 {
		p.out.Printf("  Details: %s\n", details)
	}
	p.out.Printf("  Allow? (y/n/always): ")

	resultCh := make(chan string, 1)
	go func() {
		line, _ := p.in.ReadString('\n')
		resultCh <- strings.TrimSpace(strings.ToLower(line))
	}()

	select {
	case <-ctx.Done():
		return false, false, ctx.Err()
	case answer := <-resultCh:
		switch answer {
		case "always":
			return true, true, nil
		case "y", "yes":
			return true, false, nil
		default:
			return false, false, nil
		}
	}
}
