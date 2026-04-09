// Package main is the entrypoint for the og CLI binary.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/shtdu/ohgo/internal/api"
	"github.com/shtdu/ohgo/internal/auth"
	"github.com/shtdu/ohgo/internal/bridge"
	"github.com/shtdu/ohgo/internal/commands"
	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/coordinator"
	"github.com/shtdu/ohgo/internal/engine"
	"github.com/shtdu/ohgo/internal/hooks"
	mpcpkg "github.com/shtdu/ohgo/internal/mcp"
	"github.com/shtdu/ohgo/internal/permissions"
	"github.com/shtdu/ohgo/internal/plugins"
	"github.com/shtdu/ohgo/internal/prompts"
	"github.com/shtdu/ohgo/internal/skills"
	"github.com/shtdu/ohgo/internal/tasks"
	"github.com/shtdu/ohgo/internal/tools"
	"github.com/shtdu/ohgo/internal/tools/builtin"
	toolcron "github.com/shtdu/ohgo/internal/tools/cron"
	"github.com/shtdu/ohgo/internal/tools/message"
	"github.com/shtdu/ohgo/internal/ui"
)

var (
	modelFlag    string
	profileFlag  string
	promptFlag   string
	permModeFlag string
	verboseFlag  bool
	maxTokensFlg int
)

var rootCmd = &cobra.Command{
	Use:   "og",
	Short: "OpenHarness Go - AI coding agent",
	Long:  "og is a Go reimplementation of OpenHarness, an AI-powered coding agent with tool-use, skills, memory, and permissions.",
	RunE:  run,
}

func init() {
	rootCmd.Flags().StringVarP(&modelFlag, "model", "m", "", "LLM model to use")
	rootCmd.Flags().StringVarP(&profileFlag, "profile", "p", "", "provider profile name")
	rootCmd.Flags().StringVarP(&promptFlag, "prompt", "", "", "one-shot prompt (non-interactive)")
	rootCmd.Flags().StringVarP(&permModeFlag, "permission", "", "default", "permission mode: default, plan, auto")
	rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose output")
	rootCmd.Flags().IntVarP(&maxTokensFlg, "max-tokens", "", 0, "max tokens per response")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Load config
	cfgMgr := config.NewManager("")
	cfg, err := cfgMgr.Load(ctx)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Apply CLI overrides
	if modelFlag != "" {
		cfg.Model = modelFlag
	}
	if maxTokensFlg > 0 {
		cfg.MaxTokens = maxTokensFlg
	}

	// Create provider registry and resolve API client
	apiReg := api.NewRegistry()
	apiClient, err := apiReg.CreateClient(cfg, profileFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to create API client: %v\n", err)
		apiClient = api.NewAnthropicClient() // fallback
	}

	// Wire up components
	registry := tools.NewRegistry()
	hookExec := hooks.NoopRunner()

	// Permission checker from config
	permSettings := cfg.Permission
	if permModeFlag != "" {
		permSettings.Mode = permModeFlag
	}
	permChecker := permissions.NewDefaultChecker(permSettings)

	// Interactive UI (for ask_user and permission prompts)
	termUI := ui.New(os.Stdout, os.Stdin)
	permPrompt := ui.NewPermissionPrompter(bufio.NewReader(os.Stdin), termUI)

	// Message emitter prints notifications to stderr
	msgEmitter := func(msg message.Message) error {
		prefix := "message"
		if msg.Level != "" {
			prefix = msg.Level
		}
		fmt.Fprintf(os.Stderr, "[%s] %s\n", prefix, msg.Content)
		return nil
	}

	// Register tools with dependencies
	cronMgr := toolcron.NewManager()

	// Phase 6: Subsystems
	skillReg := skills.NewRegistry()
	taskMgr := tasks.NewManager()
	pluginMgr := plugins.NewManager()
	authMgr := auth.NewManager("")

	// Config directory (used by multiple subsystems)
	skillDir, _ := config.ConfigDir()

	// Bridge subsystem
	bridgeMgr := bridge.NewManager()
	bridgeMgr.Register(bridge.NewClaudeCLI())
	bridgeMgr.Register(bridge.NewCodexBridge())
	defer bridgeMgr.CloseAll()

	// MCP subsystem
	mcpMgr := mpcpkg.NewManager()
	defer mcpMgr.CloseAll()

	// Coordinator subsystem
	coordDirs := []string{}
	if skillDir != "" {
		coordDirs = append(coordDirs, filepath.Join(skillDir, "agents"))
	}
	coordLoader := coordinator.NewLoader(coordDirs...)
	agentDefs, _ := coordLoader.LoadAll(ctx)
	ogBinPath, _ := os.Executable()
	coord := coordinator.New(ogBinPath)
	coord.RegisterDefs(agentDefs)
	defer coord.Shutdown()

	// Load user skills
	if skillDir != "" {
		loader := skills.NewLoader(filepath.Join(skillDir, "skills"))
		userSkills, err := loader.LoadAll(ctx)
		if err == nil {
			for _, s := range userSkills {
				skillReg.Register(s)
			}
		}
	}

	// Discover plugins
	pluginDirs := []string{}
	if userPluginDir := filepath.Join(skillDir, "plugins"); skillDir != "" {
		pluginDirs = append(pluginDirs, userPluginDir)
	}
	if cwd, err := os.Getwd(); err == nil {
		pluginDirs = append(pluginDirs, filepath.Join(cwd, ".openharness", "plugins"))
	}
	if len(pluginDirs) > 0 {
		if err := pluginMgr.Discover(ctx, pluginDirs...); err != nil {
			fmt.Fprintf(os.Stderr, "warning: plugin discovery failed: %v\n", err)
		}
	}

	builtin.RegisterAll(registry, builtin.ToolDeps{
		Checker:     permChecker,
		Settings:    cfg,
		Registry:    registry,
		CronMgr:     cronMgr,
		SkillReg:    skillReg,
		TaskMgr:     taskMgr,
		PluginMgr:   pluginMgr,
		MCPMgr:      mcpMgr,
		Coord:       coord,
		AskPrompter: termUI,
		MsgEmitter:  msgEmitter,
	})

	// Connect MCP servers from config.
	if len(cfg.MCP.Servers) > 0 {
		if err := mcpMgr.ConnectAll(ctx, cfg.MCP.Servers); err != nil {
			fmt.Fprintf(os.Stderr, "warning: MCP connect failed: %v\n", err)
		}
	}

	// Build system prompt
	assembler := prompts.NewAssembler("").WithCustomPrompt(cfg.SystemPrompt)
	systemPrompt, err := assembler.Build(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to build system prompt: %v\n", err)
		systemPrompt = "You are a helpful coding assistant."
	}

	// Create event channel
	eventCh := make(chan engine.EngineEvent, 64)

	// Create engine
	eng := engine.New(engine.Options{
		Model:      cfg.Model,
		MaxTokens:  cfg.MaxTokens,
		MaxTurns:   cfg.MaxTurns,
		System:     systemPrompt,
		Permission: permChecker,
		ToolReg:    registry,
		Hooks:      hookExec,
		APIClient:  apiClient,
		EventCh:    eventCh,
		PermPrompt: permPrompt,
	})

	// Register slash commands
	cmdReg := commands.NewRegistry()
	commands.RegisterAll(cmdReg)
	cwd, _ := os.Getwd()
	cmdDeps := &commands.Deps{
		Engine:    eng,
		Config:    cfg,
		ConfigMgr: cfgMgr,
		Skills:    skillReg,
		Tasks:     taskMgr,
		Plugins:   pluginMgr,
		ToolReg:   registry,
		CmdReg:    cmdReg,
		AuthMgr:   authMgr,
		BridgeMgr: bridgeMgr,
		Cwd:       cwd,
		Version:   "dev",
	}

	// Start event printer
	done := make(chan struct{})
	go func() {
		printEvents(eventCh)
		close(done)
	}()

	var queryErr error
	if promptFlag != "" {
		// One-shot mode
		queryErr = eng.Query(ctx, promptFlag)
	} else {
		// Interactive REPL
		queryErr = runREPL(ctx, eng, cmdDeps)
	}
	close(eventCh)
	<-done
	return queryErr
}

func printEvents(ch <-chan engine.EngineEvent) {
	for event := range ch {
		switch data := event.Data.(type) {
		case engine.AssistantTextDelta:
			fmt.Print(data.Text)
		case engine.ToolExecutionStarted:
			fmt.Fprintf(os.Stderr, "\n[tool: %s]\n", data.ToolName)
		case engine.ToolExecutionCompleted:
			if data.IsError {
				fmt.Fprintf(os.Stderr, "[tool error: %s]\n", data.Output)
			}
		case engine.ErrorEvent:
			fmt.Fprintf(os.Stderr, "error: %s\n", data.Message)
		case engine.AssistantTurnComplete:
			fmt.Println()
		}
	}
}

func runREPL(ctx context.Context, eng *engine.Engine, cmdDeps *commands.Deps) error {
	fmt.Println("og - OpenHarness Go")
	fmt.Println("Type a prompt and press Enter. Type /help for commands, /exit to quit.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Check for slash commands
		if cmd, args, ok := cmdDeps.CmdReg.Lookup(line); ok {
			result, err := cmd.Run(ctx, args, cmdDeps)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			} else {
				if result.Output != "" {
					fmt.Println(result.Output)
				}
				if result.ShouldExit {
					break
				}
			}
			continue
		}

		if err := eng.Query(ctx, line); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
	return scanner.Err()
}
