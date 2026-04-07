// Package main is the entrypoint for the og CLI binary.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/shtdu/ohgo/internal/api"
	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/engine"
	"github.com/shtdu/ohgo/internal/hooks"
	"github.com/shtdu/ohgo/internal/permissions"
	"github.com/shtdu/ohgo/internal/tools"
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

	// Resolve API key
	apiKey := cfg.ResolveAPIKey()
	if apiKey == "" {
		apiKey = resolveAPIKeyFromEnv()
	}

	// Wire up components
	registry := tools.NewRegistry()
	hookExec := hooks.NewExecutor()
	permChecker := permissions.NewDefaultChecker(permissions.Mode(permModeFlag))

	// Create API client
	var apiClient api.Client
	if apiKey != "" {
		apiClient = api.NewAnthropicClient(
			api.WithAPIKey(apiKey),
		)
	} else {
		fmt.Fprintf(os.Stderr, "warning: no API key configured. Set ANTHROPIC_API_KEY or configure in settings.\n")
		apiClient = api.NewAnthropicClient() // will fail on actual requests
	}

	// Create event channel
	eventCh := make(chan engine.EngineEvent, 64)

	// Create engine
	eng := engine.New(engine.Options{
		Model:      cfg.Model,
		MaxTokens:  cfg.MaxTokens,
		MaxTurns:   cfg.MaxTurns,
		System:     "You are a helpful coding assistant.",
		Permission: permChecker,
		ToolReg:    registry,
		Hooks:      hookExec,
		APIClient:  apiClient,
		EventCh:    eventCh,
	})

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
		queryErr = runREPL(ctx, eng)
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

func runREPL(ctx context.Context, eng *engine.Engine) error {
	fmt.Println("og - OpenHarness Go")
	fmt.Println("Type a prompt and press Enter. Type /exit or Ctrl+C to quit.")
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
		if line == "/exit" || line == "exit" {
			break
		}

		if err := eng.Query(ctx, line); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
	return scanner.Err()
}

func resolveAPIKeyFromEnv() string {
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key
	}
	return ""
}
