package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shtdu/ohgo/internal/config"
)

func testConfigDeps(cfg *config.Settings, cwd string) *Deps {
	if cfg == nil {
		s := config.DefaultSettings()
		cfg = &s
	}
	return &Deps{
		Config: cfg,
		Cwd:    cwd,
	}
}

func TestThemeCmd_Show(t *testing.T) {
	s := config.DefaultSettings()
	deps := testConfigDeps(&s, "/tmp")
	cmd := ThemeCmd()

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "default") {
		t.Errorf("expected output to contain 'default', got %q", res.Output)
	}
}

func TestThemeCmd_Set(t *testing.T) {
	s := config.DefaultSettings()
	deps := testConfigDeps(&s, "/tmp")
	cmd := ThemeCmd()

	res, err := cmd.Run(context.Background(), "monokai", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "monokai") {
		t.Errorf("expected output to contain 'monokai', got %q", res.Output)
	}
	if s.Theme != "monokai" {
		t.Errorf("expected Theme to be 'monokai', got %q", s.Theme)
	}
}

func TestVimCmd_Toggle(t *testing.T) {
	s := config.DefaultSettings()
	deps := testConfigDeps(&s, "/tmp")
	cmd := VimCmd()

	// First toggle: off -> on
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.VimMode {
		t.Error("expected VimMode to be true after first toggle")
	}
	if !strings.Contains(res.Output, "on") {
		t.Errorf("expected output to contain 'on', got %q", res.Output)
	}

	// Second toggle: on -> off
	res, err = cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.VimMode {
		t.Error("expected VimMode to be false after second toggle")
	}
	if !strings.Contains(res.Output, "off") {
		t.Errorf("expected output to contain 'off', got %q", res.Output)
	}
}

func TestInitCmd_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	s := config.DefaultSettings()
	deps := testConfigDeps(&s, tmpDir)
	cmd := InitCmd()

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "created") {
		t.Errorf("expected output to contain 'created', got %q", res.Output)
	}

	// Verify .ohgo directory was created.
	ohDir := filepath.Join(tmpDir, ".ohgo")
	if info, err := os.Stat(ohDir); err != nil || !info.IsDir() {
		t.Fatalf("expected .ohgo directory to exist")
	}

	// Verify settings.json was created.
	settingsPath := filepath.Join(ohDir, "settings.json")
	if _, err := os.Stat(settingsPath); err != nil {
		t.Fatalf("expected settings.json to exist: %v", err)
	}

	// Verify plugins/ directory was created.
	pluginsDir := filepath.Join(ohDir, "plugins")
	if info, err := os.Stat(pluginsDir); err != nil || !info.IsDir() {
		t.Fatalf("expected plugins/ directory to exist")
	}
}

func TestInitCmd_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	s := config.DefaultSettings()
	deps := testConfigDeps(&s, tmpDir)
	cmd := InitCmd()

	// Run twice.
	_, _ = cmd.Run(context.Background(), "", deps)
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error on second run: %v", err)
	}
	if !strings.Contains(res.Output, "already exists") {
		t.Errorf("expected output to mention 'already exists', got %q", res.Output)
	}
}

func TestFastCmd_Toggle(t *testing.T) {
	s := config.DefaultSettings()
	deps := testConfigDeps(&s, "/tmp")
	cmd := FastCmd()

	// Default: Verbose is false.
	if s.Verbose {
		t.Error("expected Verbose to start false")
	}

	// First toggle: off -> on
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.Verbose {
		t.Error("expected Verbose to be true after first toggle")
	}
	if !strings.Contains(res.Output, "on") {
		t.Errorf("expected output to contain 'on', got %q", res.Output)
	}
}

func TestEffortCmd_Valid(t *testing.T) {
	deps := testConfigDeps(nil, "/tmp")
	cmd := EffortCmd()

	// Default should be "medium".
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "medium") {
		t.Errorf("expected default to be 'medium', got %q", res.Output)
	}

	// Set to high.
	res, err = cmd.Run(context.Background(), "high", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "high") {
		t.Errorf("expected output to contain 'high', got %q", res.Output)
	}

	// Verify it persists.
	res, _ = cmd.Run(context.Background(), "", deps)
	if !strings.Contains(res.Output, "high") {
		t.Errorf("expected effort to remain 'high', got %q", res.Output)
	}
}

func TestEffortCmd_Invalid(t *testing.T) {
	deps := testConfigDeps(nil, "/tmp")
	cmd := EffortCmd()

	_, err := cmd.Run(context.Background(), "extreme", deps)
	if err == nil {
		t.Error("expected error for invalid effort value")
	}
}

func TestPassesCmd_Valid(t *testing.T) {
	deps := testConfigDeps(nil, "/tmp")
	cmd := PassesCmd()

	// Default should be 1.
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "1") {
		t.Errorf("expected default to be 1, got %q", res.Output)
	}

	// Set to 3.
	res, err = cmd.Run(context.Background(), "3", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "3") {
		t.Errorf("expected output to contain '3', got %q", res.Output)
	}
}

func TestPassesCmd_Invalid(t *testing.T) {
	deps := testConfigDeps(nil, "/tmp")
	cmd := PassesCmd()

	_, err := cmd.Run(context.Background(), "abc", deps)
	if err == nil {
		t.Error("expected error for non-integer passes")
	}

	_, err = cmd.Run(context.Background(), "0", deps)
	if err == nil {
		t.Error("expected error for passes < 1")
	}
}

func TestVoiceCmd(t *testing.T) {
	deps := testConfigDeps(nil, "/tmp")
	cmd := VoiceCmd()

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "not yet implemented") {
		t.Errorf("expected placeholder message, got %q", res.Output)
	}
}

func TestPrivacyCmd(t *testing.T) {
	deps := testConfigDeps(nil, "/tmp")
	cmd := PrivacyCmd()

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "default privacy settings active") {
		t.Errorf("expected default message, got %q", res.Output)
	}
}

func TestKeybindCmd(t *testing.T) {
	deps := testConfigDeps(nil, "/tmp")
	cmd := KeybindCmd()

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "default keybindings active") {
		t.Errorf("expected default message, got %q", res.Output)
	}
}

func TestStyleCmd_ShowAndSet(t *testing.T) {
	s := config.DefaultSettings()
	deps := testConfigDeps(&s, "/tmp")
	cmd := StyleCmd()

	// Show default.
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "default") {
		t.Errorf("expected output to contain 'default', got %q", res.Output)
	}

	// Set to compact.
	res, err = cmd.Run(context.Background(), "compact", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.OutputStyle != "compact" {
		t.Errorf("expected OutputStyle 'compact', got %q", s.OutputStyle)
	}
}
