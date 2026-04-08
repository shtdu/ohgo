// Package worktree implements git worktree management tools.
package worktree

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

type enterInput struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	CreateBranch bool   `json:"create_branch"`
}

// EnterWorktreeTool creates a new git worktree.
type EnterWorktreeTool struct{}

func (EnterWorktreeTool) Name() string { return "enter_worktree" }

func (EnterWorktreeTool) Description() string {
	return "Creates a new git worktree for isolated development. Optionally creates a new branch."
}

func (EnterWorktreeTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Name for the worktree",
			},
			"path": map[string]any{
				"type":        "string",
				"description": "Path for worktree directory. Defaults to .claude/worktrees/<name>",
			},
			"create_branch": map[string]any{
				"type":        "boolean",
				"default":     true,
				"description": "Create a new branch for the worktree",
			},
		},
		"required":             []string{"name"},
		"additionalProperties": false,
	}
}

func (EnterWorktreeTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input enterInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.Name == "" {
		return tools.Result{Content: "name is required", IsError: true}, nil
	}

	// Check context
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	// Find git root
	gitRoot, err := gitRoot(ctx)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("cannot find git root: %v", err), IsError: true}, nil
	}

	// Resolve worktree path
	wtPath := input.Path
	if wtPath == "" {
		wtPath = filepath.Join(gitRoot, ".claude", "worktrees", input.Name)
	}
	wtPath = tools.ResolvePath(wtPath)

	// Detect whether create_branch was explicitly set; default is true.
	createBranch := input.CreateBranch
	if !createBranch {
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(args, &raw)
		if _, ok := raw["create_branch"]; !ok {
			createBranch = true
		}
	}

	// Build git worktree add command
	var cmd *exec.Cmd
	if createBranch {
		cmd = exec.CommandContext(ctx, "git", "worktree", "add", "-b", input.Name, wtPath)
	} else {
		cmd = exec.CommandContext(ctx, "git", "worktree", "add", wtPath)
	}
	cmd.Dir = gitRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return tools.Result{
			Content: fmt.Sprintf("git worktree add failed: %s", strings.TrimSpace(string(output))),
			IsError: true,
		}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Created worktree at %s", wtPath)}, nil
}

func gitRoot(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
