// Package bash implements the bash tool for executing shell commands.
package bash

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/shtdu/ohgo/internal/tools"
)

const (
	defaultTimeout  = 120
	maxTimeout      = 600
	maxOutput       = 12000
	truncatedSuffix = "\n...[truncated]..."
)

type bashInput struct {
	Command        string `json:"command"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// BashTool executes shell commands via sh -c.
type BashTool struct{}

func (BashTool) Name() string { return "bash" }

func (BashTool) Description() string {
	return "Executes a bash command in a shell. Returns combined stdout and stderr output."
}

func (BashTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "The bash command to execute",
			},
			"timeout_seconds": map[string]any{
				"type":        "integer",
				"description": "Optional timeout in seconds (max 600)",
				"default":     defaultTimeout,
				"minimum":     1,
				"maximum":     maxTimeout,
			},
		},
		"required":             []string{"command"},
		"additionalProperties": false,
	}
}

func (BashTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input bashInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{
			Content: fmt.Sprintf("invalid arguments: %v", err),
			IsError: true,
		}, nil
	}

	if input.Command == "" {
		return tools.Result{
			Content: "command is required",
			IsError: true,
		}, nil
	}

	timeout := input.TimeoutSeconds
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if timeout > maxTimeout {
		timeout = maxTimeout
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.Command("sh", "-c", input.Command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Use a pipe for stdout and write stderr to the same pipe via a multi-writer.
	// Actually, use a simpler approach: set Stdout and Stderr to the same buffer.
	var combinedBuf bytes.Buffer
	cmd.Stdout = &combinedBuf
	cmd.Stderr = &combinedBuf

	if err := cmd.Start(); err != nil {
		return tools.Result{
			Content: fmt.Sprintf("failed to start command: %v", err),
			IsError: true,
		}, nil
	}

	// Kill the process group when the context is done.
	go func() {
		<-timeoutCtx.Done()
		// Send SIGTERM first for graceful shutdown.
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err == nil {
			// Give the process a brief grace period before force-killing.
			time.Sleep(2 * time.Second)
		}
		// The process may have already exited; ignore the error.
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}()

	waitErr := cmd.Wait()

	output := combinedBuf.String()

	// Determine the failure reason.
	if timeoutCtx.Err() == context.DeadlineExceeded {
		return tools.Result{
			Content: fmt.Sprintf("[command timed out after %ds]", timeout),
			IsError: true,
		}, nil
	}

	if ctx.Err() != nil {
		return tools.Result{}, ctx.Err()
	}

	out := truncateOutput(output)

	if waitErr != nil {
		exitCode := exitCodeFromError(waitErr)
		if out != "" && !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		out += fmt.Sprintf("[exit code: %d]", exitCode)
		return tools.Result{
			Content: out,
			IsError: true,
		}, nil
	}

	return tools.Result{Content: out}, nil
}

func truncateOutput(s string) string {
	if len(s) <= maxOutput {
		return s
	}
	return s[:maxOutput] + truncatedSuffix
}

func exitCodeFromError(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 1
}
