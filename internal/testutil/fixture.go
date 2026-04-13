//go:build integration

package testutil

import (
	"testing"

	"github.com/shtdu/ohgo/internal/api"
	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/engine"
	"github.com/shtdu/ohgo/internal/hooks"
	"github.com/shtdu/ohgo/internal/permissions"
	"github.com/shtdu/ohgo/internal/tools"
	"github.com/shtdu/ohgo/internal/tools/bash"
	"github.com/shtdu/ohgo/internal/tools/edit"
	"github.com/shtdu/ohgo/internal/tools/glob"
	"github.com/shtdu/ohgo/internal/tools/grep"
	"github.com/shtdu/ohgo/internal/tools/read"
	"github.com/shtdu/ohgo/internal/tools/write"
)

// Fixture provides a fully wired integration test environment.
// All components are real except the API client (mocked) and user prompts (mocked).
type Fixture struct {
	Dir      string                       // temp project directory
	Registry *tools.Registry              // real tool registry
	Checker  *permissions.DefaultChecker  // real permission checker
	Engine   *engine.Engine               // wired engine
	MockAPI  *MockAPIClient               // mock API client
	MockHooks *MockHookRunner             // mock hooks
	Prompter *MockPermissionPrompter      // mock prompter
	EventCh  chan engine.EngineEvent
}

// FixtureConfig customizes the test fixture.
type FixtureConfig struct {
	// Responses scripts the mock API client responses.
	Responses []ResponseScript
	// PermMode sets the initial permission mode (default: "default").
	PermMode string
	// AllowedTools sets tools on the explicit allow list.
	AllowedTools []string
	// DeniedTools sets tools on the explicit deny list.
	DeniedTools []string
	// PathRules sets file path permission rules.
	PathRules []config.PathRuleConfig
	// MaxTurns sets the engine max turns (default: 10).
	MaxTurns int
	// Model sets the engine model name.
	Model string
	// System sets the system prompt.
	System string
}

// NewFixture creates a fully wired test fixture.
func NewFixture(t *testing.T, cfg FixtureConfig) *Fixture {
	t.Helper()

	dir := TempDir(t)

	// Build tool registry with real file tools.
	registry := tools.NewRegistry()
	registry.Register(bash.BashTool{})
	registry.Register(read.ReadTool{})
	registry.Register(write.WriteTool{})
	registry.Register(edit.EditTool{})
	registry.Register(glob.GlobTool{})
	registry.Register(grep.GrepTool{})


	// Permission checker.
	permMode := cfg.PermMode
	if permMode == "" {
		permMode = "default"
	}
	checker := permissions.NewDefaultChecker(config.PermissionSettings{
		Mode:         permMode,
		AllowedTools: cfg.AllowedTools,
		DeniedTools:  cfg.DeniedTools,
		PathRules:    cfg.PathRules,
	})

	// Mock API client.
	mockAPI := NewMockAPIClient(cfg.Responses...)

	// Mock hooks (non-blocking by default).
	mockHooks := NewMockHookRunner(false, "")

	// Mock prompter (deny by default — tests opt in to allow).
	prompter := NewMockPermissionPrompter(
		PromptResponse{Allow: false},
		nil,
	)

	// Event channel.
	eventCh := make(chan engine.EngineEvent, 64)

	// Engine.
	maxTurns := cfg.MaxTurns
	if maxTurns == 0 {
		maxTurns = 10
	}
	eng := engine.New(engine.Options{
		Model:      cfg.Model,
		MaxTurns:   maxTurns,
		System:     cfg.System,
		Permission: checker,
		ToolReg:    registry,
		Hooks:      mockHooks,
		APIClient:  mockAPI,
		EventCh:    eventCh,
		PermPrompt: prompter,
	})

	return &Fixture{
		Dir:       dir,
		Registry:  registry,
		Checker:   checker,
		Engine:    eng,
		MockAPI:   mockAPI,
		MockHooks: mockHooks,
		Prompter:  prompter,
		EventCh:   eventCh,
	}
}

// DrainEvents drains the event channel in a background goroutine.
// Returns collected events and a stop function.
func (f *Fixture) DrainEvents() (events []engine.EngineEvent, stop func() []engine.EngineEvent) {
	var collected []engine.EngineEvent
	done := make(chan struct{})
	go func() {
		defer close(done)
		for e := range f.EventCh {
			collected = append(collected, e)
		}
	}()
	return collected, func() []engine.EngineEvent {
		close(f.EventCh)
		<-done
		return collected
	}
}

// Compile-time interface checks.
var _ hooks.HookRunner = (*MockHookRunner)(nil)
var _ engine.PermissionPrompter = (*MockPermissionPrompter)(nil)
var _ api.Client = (*MockAPIClient)(nil)
