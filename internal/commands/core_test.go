package commands

import (
	"context"
	"strings"
	"testing"

	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/engine"
)

func TestExitCommand(t *testing.T) {
	cmd := exitCmd{}
	deps := &Deps{}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("exit should not error: %v", err)
	}
	if !res.ShouldExit {
		t.Error("exit should set ShouldExit")
	}
	if res.Output == "" {
		t.Error("exit should produce output")
	}
}

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{"shows configured version", "1.2.3", "1.2.3"},
		{"falls back to dev", "", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := versionCmd{}
			deps := &Deps{Version: tt.version}
			res, err := cmd.Run(context.Background(), "", deps)
			if err != nil {
				t.Fatalf("version should not error: %v", err)
			}
			if res.Output != tt.want {
				t.Errorf("version output = %q, want %q", res.Output, tt.want)
			}
		})
	}
}

func TestClearCommand(t *testing.T) {
	eng := engine.New(engine.Options{Model: "test-model"})
	cmd := clearCmd{}
	deps := &Deps{Engine: eng}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("clear should not error: %v", err)
	}
	if res.Output != "Conversation cleared." {
		t.Errorf("clear output = %q, want %q", res.Output, "Conversation cleared.")
	}
}

func TestStatusCommand(t *testing.T) {
	eng := engine.New(engine.Options{Model: "test-model"})
	cmd := statusCmd{}
	deps := &Deps{Engine: eng}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("status should not error: %v", err)
	}
	if !strings.Contains(res.Output, "test-model") {
		t.Errorf("status output should contain model name, got: %s", res.Output)
	}
	if !strings.Contains(res.Output, "Turns:") {
		t.Errorf("status output should contain Turns, got: %s", res.Output)
	}
}

func TestCostCommand(t *testing.T) {
	eng := engine.New(engine.Options{Model: "test-model"})
	cmd := costCmd{}
	deps := &Deps{Engine: eng}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("cost should not error: %v", err)
	}
	if !strings.Contains(res.Output, "Total tokens:") {
		t.Errorf("cost output should contain total tokens, got: %s", res.Output)
	}
}

func TestUsageCommand(t *testing.T) {
	eng := engine.New(engine.Options{Model: "test-model"})
	cmd := usageCmd{}
	deps := &Deps{Engine: eng}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("usage should not error: %v", err)
	}
	if !strings.Contains(res.Output, "tokens") {
		t.Errorf("usage output should mention tokens, got: %s", res.Output)
	}
}

func TestStatsCommand(t *testing.T) {
	eng := engine.New(engine.Options{Model: "test-model"})
	cmd := statsCmd{}
	deps := &Deps{Engine: eng}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("stats should not error: %v", err)
	}
	if !strings.Contains(res.Output, "Turns:") {
		t.Errorf("stats output should contain Turns, got: %s", res.Output)
	}
	if !strings.Contains(res.Output, "Messages:") {
		t.Errorf("stats output should contain Messages, got: %s", res.Output)
	}
}

func TestHelpCommand(t *testing.T) {
	reg := NewRegistry()
	registerTestCore(reg)

	cmd := helpCmd{}
	deps := &Deps{CmdReg: reg}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("help should not error: %v", err)
	}
	if !strings.Contains(res.Output, "/help") {
		t.Errorf("help output should list /help, got: %s", res.Output)
	}
	if !strings.Contains(res.Output, "/exit") {
		t.Errorf("help output should list /exit, got: %s", res.Output)
	}
}

func TestHelpCommand_NoRegistry(t *testing.T) {
	cmd := helpCmd{}
	deps := &Deps{CmdReg: nil}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("help should not error: %v", err)
	}
	if !strings.Contains(res.Output, "No commands") {
		t.Errorf("help with nil registry should say no commands, got: %s", res.Output)
	}
}

func TestModelCommand_Show(t *testing.T) {
	eng := engine.New(engine.Options{Model: "claude-test"})
	cmd := modelCmd{}
	deps := &Deps{Engine: eng}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("model show should not error: %v", err)
	}
	if !strings.Contains(res.Output, "claude-test") {
		t.Errorf("model output should contain model name, got: %s", res.Output)
	}
}

func TestModelCommand_Switch(t *testing.T) {
	eng := engine.New(engine.Options{Model: "old-model"})
	cmd := modelCmd{}
	deps := &Deps{Engine: eng}

	res, err := cmd.Run(context.Background(), "new-model", deps)
	if err != nil {
		t.Fatalf("model switch should not error: %v", err)
	}
	if eng.Model() != "new-model" {
		t.Errorf("engine model = %q, want %q", eng.Model(), "new-model")
	}
	if !strings.Contains(res.Output, "old-model") || !strings.Contains(res.Output, "new-model") {
		t.Errorf("model switch output should mention both models, got: %s", res.Output)
	}
}

func TestProviderCommand_Show(t *testing.T) {
	cfg := config.DefaultSettings()
	cmd := providerCmd{}
	deps := &Deps{Config: &cfg, Engine: engine.New(engine.Options{})}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("provider show should not error: %v", err)
	}
	if !strings.Contains(res.Output, "claude-api") {
		t.Errorf("provider output should contain default profile, got: %s", res.Output)
	}
}

func TestProviderCommand_Switch(t *testing.T) {
	cfg := config.DefaultSettings()
	eng := engine.New(engine.Options{Model: "test"})
	cmd := providerCmd{}
	deps := &Deps{Config: &cfg, Engine: eng}

	_, err := cmd.Run(context.Background(), "openai-compatible", deps)
	if err != nil {
		t.Fatalf("provider switch should not error: %v", err)
	}
	if cfg.ActiveProfile != "openai-compatible" {
		t.Errorf("active profile = %q, want %q", cfg.ActiveProfile, "openai-compatible")
	}
}

func TestProviderCommand_UnknownProfile(t *testing.T) {
	cfg := config.DefaultSettings()
	cmd := providerCmd{}
	deps := &Deps{Config: &cfg, Engine: engine.New(engine.Options{})}

	_, err := cmd.Run(context.Background(), "nonexistent", deps)
	if err == nil {
		t.Fatal("provider switch with unknown profile should error")
	}
}

func TestPermissionsCommand(t *testing.T) {
	cfg := config.DefaultSettings()
	cmd := permissionsCmd{}
	deps := &Deps{Config: &cfg}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("permissions should not error: %v", err)
	}
	if !strings.Contains(res.Output, "default") {
		t.Errorf("permissions output should contain mode, got: %s", res.Output)
	}
}

func TestPlanCommand_Toggle(t *testing.T) {
	cfg := config.DefaultSettings()

	cmd := planCmd{}

	// Toggle on
	deps := &Deps{Config: &cfg}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("plan toggle should not error: %v", err)
	}
	if cfg.OutputStyle != "plan" {
		t.Errorf("output style = %q, want %q", cfg.OutputStyle, "plan")
	}
	if !strings.Contains(res.Output, "enabled") {
		t.Errorf("plan output should say enabled, got: %s", res.Output)
	}

	// Toggle off
	res, err = cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("plan toggle off should not error: %v", err)
	}
	if cfg.OutputStyle != "default" {
		t.Errorf("output style = %q, want %q", cfg.OutputStyle, "default")
	}
}

func TestPlanCommand_ExplicitArgs(t *testing.T) {
	cfg := config.DefaultSettings()
	cmd := planCmd{}

	// Enable with "on"
	deps := &Deps{Config: &cfg}
	_, err := cmd.Run(context.Background(), "on", deps)
	if err != nil {
		t.Fatalf("plan on should not error: %v", err)
	}
	if cfg.OutputStyle != "plan" {
		t.Errorf("output style = %q, want %q", cfg.OutputStyle, "plan")
	}

	// Disable with "off"
	res2, err := cmd.Run(context.Background(), "off", deps)
	if err != nil {
		t.Fatalf("plan off should not error: %v", err)
	}
	if cfg.OutputStyle != "default" {
		t.Errorf("output style = %q, want %q", cfg.OutputStyle, "default")
	}
	if !strings.Contains(res2.Output, "disabled") {
		t.Errorf("plan output should say disabled, got: %s", res2.Output)
	}
}

func TestPlanCommand_BadArg(t *testing.T) {
	cfg := config.DefaultSettings()
	cmd := planCmd{}
	deps := &Deps{Config: &cfg}

	_, err := cmd.Run(context.Background(), "badval", deps)
	if err == nil {
		t.Fatal("plan with bad arg should error")
	}
}

func TestRegisterCore(t *testing.T) {
	reg := NewRegistry()
	registerTestCore(reg)

	expected := []string{
		"exit", "version", "clear", "status", "cost",
		"usage", "stats", "help", "model", "provider",
		"permissions", "plan",
	}

	for _, name := range expected {
		cmd := reg.Get(name)
		if cmd == nil {
			t.Errorf("expected command %q to be registered", name)
			continue
		}
		if cmd.Name() != name {
			t.Errorf("command name = %q, want %q", cmd.Name(), name)
		}
		if cmd.ShortHelp() == "" {
			t.Errorf("command %q should have non-empty ShortHelp", name)
		}
	}
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry()
	registerTestCore(reg)

	cmds := reg.List()
	if len(cmds) != 12 {
		t.Fatalf("expected 12 commands, got %d", len(cmds))
	}

	// Verify sorted order
	for i := 1; i < len(cmds); i++ {
		if cmds[i].Name() < cmds[i-1].Name() {
			t.Errorf("commands not sorted: %q before %q", cmds[i-1].Name(), cmds[i].Name())
		}
	}
}

// registerTestCore adds the core commands for testing.
func registerTestCore(r *Registry) {
	r.Register(exitCmd{})
	r.Register(versionCmd{})
	r.Register(clearCmd{})
	r.Register(statusCmd{})
	r.Register(costCmd{})
	r.Register(usageCmd{})
	r.Register(statsCmd{})
	r.Register(helpCmd{})
	r.Register(modelCmd{})
	r.Register(providerCmd{})
	r.Register(permissionsCmd{})
	r.Register(planCmd{})
}
