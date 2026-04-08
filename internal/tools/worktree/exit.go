package worktree

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

type exitInput struct {
	Path          string `json:"path"`
	Action        string `json:"action"`
	DiscardChanges bool  `json:"discard_changes"`
}

// ExitWorktreeTool removes a git worktree.
type ExitWorktreeTool struct{}

func (ExitWorktreeTool) Name() string { return "exit_worktree" }

func (ExitWorktreeTool) Description() string {
	return "Removes a git worktree. Can keep or remove the worktree directory."
}

func (ExitWorktreeTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the worktree to remove",
			},
			"action": map[string]any{
				"type":        "string",
				"enum":        []string{"keep", "remove"},
				"description": "Keep or remove the worktree",
			},
			"discard_changes": map[string]any{
				"type":        "boolean",
				"default":     false,
				"description": "Force removal even with uncommitted changes",
			},
		},
		"required":             []string{"path", "action"},
		"additionalProperties": false,
	}
}

func (ExitWorktreeTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input exitInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.Path == "" {
		return tools.Result{Content: "path is required", IsError: true}, nil
	}

	// Validate action
	switch input.Action {
	case "keep", "remove":
		// valid
	default:
		return tools.Result{
			Content: fmt.Sprintf("invalid action %q: must be keep or remove", input.Action),
			IsError: true,
		}, nil
	}

	// Check context
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	if input.Action == "keep" {
		return tools.Result{Content: fmt.Sprintf("Worktree kept at %s", input.Path)}, nil
	}

	// Remove the worktree
	path := tools.ResolvePath(input.Path)

	cmdArgs := []string{"worktree", "remove"}
	if input.DiscardChanges {
		cmdArgs = append(cmdArgs, "--force")
	}
	cmdArgs = append(cmdArgs, path)

	// Find git root so the command runs from the main repo
	gitRoot, err := gitRoot(ctx)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("cannot find git root: %v", err), IsError: true}, nil
	}

	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	cmd.Dir = gitRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return tools.Result{
			Content: fmt.Sprintf("git worktree remove failed: %s", strings.TrimSpace(string(output))),
			IsError: true,
		}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Removed worktree at %s", path)}, nil
}
