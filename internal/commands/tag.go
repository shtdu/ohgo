package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// tagCmd saves or lists named snapshots of the conversation.
type tagCmd struct{}

var _ Command = tagCmd{}

func (tagCmd) Name() string     { return "tag" }
func (tagCmd) ShortHelp() string { return "save or list named conversation snapshots" }

func (tagCmd) Run(_ context.Context, args string, deps *Deps) (Result, error) {
	args = strings.TrimSpace(args)

	// No args: list tags
	if args == "" {
		return listTags()
	}

	// Save snapshot with the given tag name
	return saveTag(args, deps)
}

func listTags() (Result, error) {
	dir := sessionDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{Output: "tag: no saved tags"}, nil
		}
		return Result{}, fmt.Errorf("tag: read directory: %w", err)
	}

	if len(entries) == 0 {
		return Result{Output: "tag: no saved tags"}, nil
	}

	out := "Saved tags:\n"
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		out += fmt.Sprintf("  %s\n", name)
	}

	return Result{Output: out}, nil
}

func saveTag(tag string, deps *Deps) (Result, error) {
	msgs := deps.Engine.Messages()
	if len(msgs) == 0 {
		return Result{Output: "tag: no messages to save"}, nil
	}

	dir := sessionDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Result{}, fmt.Errorf("tag: create session dir: %w", err)
	}

	data, err := json.MarshalIndent(msgs, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("tag: marshal: %w", err)
	}

	path := filepath.Join(dir, tag+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return Result{}, fmt.Errorf("tag: write: %w", err)
	}

	return Result{
		Output: fmt.Sprintf("tag: saved %d messages as %q", len(msgs), tag),
	}, nil
}
