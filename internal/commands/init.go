package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// initCmd initializes project configuration files.
type initCmd struct{}

var _ Command = initCmd{}

func (initCmd) Name() string      { return "init" }
func (initCmd) ShortHelp() string { return "Initialize .ohgo/ project directory" }

func (initCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	dir := filepath.Join(deps.Cwd, ".ohgo")

	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return Result{Output: fmt.Sprintf("init: %s already exists", dir)}, nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Result{}, fmt.Errorf("init: create directory: %w", err)
	}

	// Write default settings.json.
	settings := map[string]interface{}{
		"model":        "claude-sonnet-4-6",
		"max_tokens":   16384,
		"api_format":   "anthropic",
		"theme":        "default",
		"output_style": "default",
		"permission": map[string]interface{}{
			"mode": "default",
		},
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("init: marshal settings: %w", err)
	}
	data = append(data, '\n')

	settingsPath := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
		return Result{}, fmt.Errorf("init: write settings: %w", err)
	}

	// Create empty plugins directory.
	pluginsDir := filepath.Join(dir, "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("init: create plugins dir: %w", err)
	}

	return Result{Output: fmt.Sprintf("init: created %s with default settings", dir)}, nil
}

// InitCmd returns a new init command.
func InitCmd() Command { return initCmd{} }
