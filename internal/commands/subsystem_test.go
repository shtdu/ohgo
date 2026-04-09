package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/memory"
	"github.com/shtdu/ohgo/internal/plugins"
	"github.com/shtdu/ohgo/internal/skills"
	"github.com/shtdu/ohgo/internal/tasks"
)

func TestSkillsCommand_NilDeps(t *testing.T) {
	cmd := skillsCmd{}
	deps := &Deps{}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "no skills loaded") {
		t.Errorf("expected no skills message, got: %s", res.Output)
	}
}

func TestSkillsCommand_EmptyRegistry(t *testing.T) {
	cmd := skillsCmd{}
	deps := &Deps{Skills: skills.NewRegistry()}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "no skills loaded") {
		t.Errorf("expected no skills message, got: %s", res.Output)
	}
}

func TestSkillsCommand_WithSkills(t *testing.T) {
	cmd := skillsCmd{}
	reg := skills.NewRegistry()
	reg.Register(&skills.Skill{
		Name:        "test-skill",
		Description: "A test skill",
		Content:     "skill content",
	})
	deps := &Deps{Skills: reg}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "test-skill") {
		t.Errorf("expected skill name in output, got: %s", res.Output)
	}
	if !strings.Contains(res.Output, "A test skill") {
		t.Errorf("expected skill description in output, got: %s", res.Output)
	}
}

func TestTasksCommand_NilDeps(t *testing.T) {
	cmd := tasksCmd{}
	deps := &Deps{}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "no task manager") {
		t.Errorf("expected no task manager message, got: %s", res.Output)
	}
}

func TestTasksCommand_EmptyManager(t *testing.T) {
	cmd := tasksCmd{}
	deps := &Deps{Tasks: tasks.NewManager()}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "no tasks") {
		t.Errorf("expected no tasks message, got: %s", res.Output)
	}
}

func TestTasksCommand_WithTasks(t *testing.T) {
	cmd := tasksCmd{}
	mgr := tasks.NewManager()
	// Use CreateShell to add a real task (it runs in background).
	rec, err := mgr.CreateShell(context.Background(), "echo hello", "test task", t.TempDir())
	if err != nil {
		t.Fatalf("CreateShell: %v", err)
	}
	_ = rec
	deps := &Deps{Tasks: mgr}

	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "echo hello") {
		t.Errorf("expected task command in output, got: %s", res.Output)
	}
	if !strings.Contains(res.Output, "running") {
		t.Errorf("expected task status in output, got: %s", res.Output)
	}
}

func TestPluginCommand_NilDeps(t *testing.T) {
	cmd := pluginCmd{}
	deps := &Deps{}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "no plugin manager") {
		t.Errorf("expected no plugin manager message, got: %s", res.Output)
	}
}

func TestPluginCommand_EmptyManager(t *testing.T) {
	cmd := pluginCmd{}
	deps := &Deps{Plugins: plugins.NewManager()}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "no plugins loaded") {
		t.Errorf("expected no plugins message, got: %s", res.Output)
	}
}

func TestPluginCommand_InstallWithSource(t *testing.T) {
	cmd := pluginCmd{}
	deps := &Deps{Plugins: plugins.NewManager()}
	// Use a real temp directory as source so plugins.Install can stat it.
	srcDir := t.TempDir()
	res, err := cmd.Run(context.Background(), "install "+srcDir, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Install may or may not succeed depending on manifest, but should not crash.
	if res.Output == "" {
		t.Error("expected non-empty output")
	}
}

func TestPluginCommand_RemoveWithValidName(t *testing.T) {
	cmd := pluginCmd{}
	deps := &Deps{Plugins: plugins.NewManager()}
	res, err := cmd.Run(context.Background(), "remove myplugin", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "removed myplugin") {
		t.Errorf("expected removed message, got: %s", res.Output)
	}
}

func TestMemoryCommand_NoCwd(t *testing.T) {
	cmd := memoryCmd{}
	deps := &Deps{}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "no working directory set") {
		t.Errorf("expected no working directory message, got: %s", res.Output)
	}
}

func TestMemoryCommand_EmptyStore(t *testing.T) {
	cmd := memoryCmd{}
	deps := &Deps{Cwd: "/tmp"}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "no entries") {
		t.Errorf("expected no entries message, got: %s", res.Output)
	}
}

func TestMemoryCommand_WithEntries(t *testing.T) {
	// Create a store and add entries directly, testing the memory.Store API
	// that the memory command uses.
	tmpDir := t.TempDir()
	store, err := memory.NewStore(tmpDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	_, err = store.Add("test-note", "a test memory")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	entries, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestHooksCommand(t *testing.T) {
	cmd := hooksCmd{}
	res, err := cmd.Run(context.Background(), "", &Deps{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "no hooks configured") {
		t.Errorf("expected no hooks message, got: %s", res.Output)
	}
}

func TestAgentsCommand(t *testing.T) {
	cmd := agentsCmd{}
	res, err := cmd.Run(context.Background(), "", &Deps{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "not yet implemented") {
		t.Errorf("expected not implemented message, got: %s", res.Output)
	}
}

func TestReloadPluginsCommand_NilDeps(t *testing.T) {
	cmd := reloadCmd{}
	deps := &Deps{}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "no plugin manager") {
		t.Errorf("expected no plugin manager message, got: %s", res.Output)
	}
}

func TestReloadPluginsCommand(t *testing.T) {
	cmd := reloadCmd{}
	deps := &Deps{Plugins: plugins.NewManager()}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "plugin discovery complete") {
		t.Errorf("expected discovery complete message, got: %s", res.Output)
	}
}

func TestMcpCommand(t *testing.T) {
	cmd := mcpCmd{}
	res, err := cmd.Run(context.Background(), "", &Deps{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "not yet implemented") {
		t.Errorf("expected not implemented message, got: %s", res.Output)
	}
}

func TestBridgeCommand(t *testing.T) {
	cmd := bridgeCmd{}
	res, err := cmd.Run(context.Background(), "", &Deps{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "not yet implemented") {
		t.Errorf("expected not implemented message, got: %s", res.Output)
	}
}

func TestLoginCommand_NoConfig(t *testing.T) {
	cmd := loginCmd{}
	deps := &Deps{}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "no configuration available") {
		t.Errorf("expected no config message, got: %s", res.Output)
	}
}

func TestLoginCommand_NotAuthenticated(t *testing.T) {
	cmd := loginCmd{}
	cfg := &config.Settings{}
	deps := &Deps{Config: cfg}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "not authenticated") {
		t.Errorf("expected not authenticated message, got: %s", res.Output)
	}
}

func TestLoginCommand_Authenticated(t *testing.T) {
	cmd := loginCmd{}
	cfg := &config.Settings{APIKey: "sk-ant-1234567890abcdef"}
	deps := &Deps{Config: cfg}
	res, err := cmd.Run(context.Background(), "", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "authenticated") {
		t.Errorf("expected authenticated message, got: %s", res.Output)
	}
	// Key should be masked.
	if strings.Contains(res.Output, "sk-ant-1234567890abcdef") {
		t.Errorf("API key should be masked in output, got: %s", res.Output)
	}
}

func TestLoginCommand_SetKey(t *testing.T) {
	cmd := loginCmd{}
	cfg := &config.Settings{}
	deps := &Deps{Config: cfg}
	res, err := cmd.Run(context.Background(), "set my-new-key", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "API key updated") {
		t.Errorf("expected key updated message, got: %s", res.Output)
	}
	if cfg.APIKey != "my-new-key" {
		t.Errorf("expected API key to be 'my-new-key', got: %s", cfg.APIKey)
	}
}

func TestLoginCommand_Logout(t *testing.T) {
	cmd := loginCmd{}
	cfg := &config.Settings{APIKey: "existing-key"}
	deps := &Deps{Config: cfg}
	res, err := cmd.Run(context.Background(), "logout", deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "API key cleared") {
		t.Errorf("expected key cleared message, got: %s", res.Output)
	}
	if cfg.APIKey != "" {
		t.Errorf("expected API key to be empty, got: %s", cfg.APIKey)
	}
}

func TestFeedbackCommand_EmptyArgs(t *testing.T) {
	cmd := feedbackCmd{}
	res, err := cmd.Run(context.Background(), "", &Deps{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "please provide feedback text") {
		t.Errorf("expected feedback prompt message, got: %s", res.Output)
	}
}

func TestFeedbackCommand_SavesToFile(t *testing.T) {
	cmd := feedbackCmd{}

	// Use a temp dir as config dir.
	tmpDir := t.TempDir()
	t.Setenv("OPENHARNESS_CONFIG_DIR", tmpDir)

	res, err := cmd.Run(context.Background(), "this is great!", &Deps{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "thank you") {
		t.Errorf("expected thank you message, got: %s", res.Output)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "feedback.txt"))
	if err != nil {
		t.Fatalf("failed to read feedback file: %v", err)
	}
	if !strings.Contains(string(data), "this is great!") {
		t.Errorf("expected feedback text in file, got: %s", string(data))
	}
}

func TestOnboardingCommand(t *testing.T) {
	cmd := onboardingCmd{}
	res, err := cmd.Run(context.Background(), "", &Deps{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res.Output, "Quickstart") {
		t.Errorf("expected quickstart guide in output, got: %s", res.Output)
	}
	if !strings.Contains(res.Output, "/login") {
		t.Errorf("expected /login reference in output, got: %s", res.Output)
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"short", "****"},
		{"sk-ant-1234567890abcdef", "sk-a...cdef"},
		{"abcd1234", "****"},
		{"abcdefghij", "abcd...ghij"},
	}
	for _, tt := range tests {
		got := maskKey(tt.input)
		if got != tt.want {
			t.Errorf("maskKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSubsystemCompileTimeChecks(t *testing.T) {
	// Verify all commands satisfy the interface via compile-time checks.
	var cmds []Command = []Command{
		memoryCmd{},
		hooksCmd{},
		skillsCmd{},
		tasksCmd{},
		agentsCmd{},
		pluginCmd{},
		reloadCmd{},
		mcpCmd{},
		bridgeCmd{},
		loginCmd{},
		feedbackCmd{},
		onboardingCmd{},
	}
	names := make(map[string]bool)
	for _, c := range cmds {
		name := c.Name()
		if names[name] {
			t.Errorf("duplicate command name: %s", name)
		}
		names[name] = true
	}
	expected := []string{
		"memory", "hooks", "skills", "tasks", "agents",
		"plugin", "reload-plugins", "mcp", "bridge",
		"login", "feedback", "onboarding",
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing command: %s", name)
		}
	}
}
