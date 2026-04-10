package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// exportCmd exports the conversation as JSON to a temp file.
type exportCmd struct{}

var _ Command = exportCmd{}

func (exportCmd) Name() string     { return "export" }
func (exportCmd) ShortHelp() string { return "export conversation as JSON" }

func (exportCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	msgs := deps.Engine.Messages()
	if len(msgs) == 0 {
		return Result{Output: "export: no messages to export"}, nil
	}

	data, err := json.MarshalIndent(msgs, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("export: marshal: %w", err)
	}

	dir := os.TempDir()
	f, err := os.CreateTemp(dir, "ohgo-export-*.json")
	if err != nil {
		return Result{}, fmt.Errorf("export: create temp file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		return Result{}, fmt.Errorf("export: write: %w", err)
	}

	return Result{
		Output: fmt.Sprintf("export: conversation saved to %s", f.Name()),
	}, nil
}

// sessionDir returns the directory used for session snapshots.
// Uses the user's config directory to avoid collisions between processes.
func sessionDir() string {
	if cfgDir, err := configDir(); err == nil && cfgDir != "" {
		return filepath.Join(cfgDir, "sessions")
	}
	return filepath.Join(os.TempDir(), "ohgo-sessions")
}

// configDir returns the user's ohgo config directory.
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".openharness"), nil
}
