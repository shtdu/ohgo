// Package commands implements slash commands (/help, /commit, /plan, etc.).
package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/shtdu/ohgo/internal/auth"
	"github.com/shtdu/ohgo/internal/bridge"
	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/engine"
	"github.com/shtdu/ohgo/internal/plugins"
	"github.com/shtdu/ohgo/internal/skills"
	"github.com/shtdu/ohgo/internal/tasks"
	"github.com/shtdu/ohgo/internal/tools"
)

// Result is what a slash command returns.
type Result struct {
	Output      string
	ShouldExit  bool
	ClearScreen bool
}

// Deps holds shared dependencies that commands can access.
type Deps struct {
	Engine    *engine.Engine
	Config    *config.Settings
	ConfigMgr *config.Manager
	Skills    *skills.Registry
	Tasks     *tasks.Manager
	Plugins   *plugins.Manager
	ToolReg   *tools.Registry
	CmdReg    *Registry
	AuthMgr   *auth.Manager
	BridgeMgr *bridge.Manager
	Cwd       string
	Version   string
}

// Command represents a slash command handler.
type Command interface {
	// Name returns the command name (e.g. "help", "commit").
	Name() string

	// ShortHelp returns a one-line description for /help listings.
	ShortHelp() string

	// Run executes the command with the given arguments and dependencies.
	Run(ctx context.Context, args string, deps *Deps) (Result, error)
}

// Registry manages available slash commands.
type Registry struct {
	commands map[string]Command
}

// NewRegistry creates a new command registry.
func NewRegistry() *Registry {
	return &Registry{commands: make(map[string]Command)}
}

// Register adds a command.
func (r *Registry) Register(c Command) {
	r.commands[c.Name()] = c
}

// Get retrieves a command by name.
func (r *Registry) Get(name string) Command {
	return r.commands[name]
}

// Lookup parses a slash command line ("/name args") and returns the
// command and remaining arguments. Returns false if the line is not
// a slash command or the command is unknown.
func (r *Registry) Lookup(line string) (Command, string, bool) {
	if !strings.HasPrefix(line, "/") {
		return nil, "", false
	}
	// Split "/name" from the rest.
	rest := line[1:]
	name, args, _ := strings.Cut(rest, " ")
	name = strings.ToLower(name)
	cmd := r.commands[name]
	if cmd == nil {
		return nil, "", false
	}
	return cmd, args, true
}

// List returns all registered commands sorted by name.
func (r *Registry) List() []Command {
	list := make([]Command, 0, len(r.commands))
	for _, c := range r.commands {
		list = append(list, c)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})
	return list
}

// formatKV formats key-value pairs into aligned lines.
func formatKV(pairs ...string) string {
	if len(pairs)%2 != 0 {
		panic("formatKV requires even number of arguments")
	}
	maxKey := 0
	for i := 0; i < len(pairs); i += 2 {
		if len(pairs[i]) > maxKey {
			maxKey = len(pairs[i])
		}
	}
	var b strings.Builder
	for i := 0; i < len(pairs); i += 2 {
		fmt.Fprintf(&b, "%-*s  %s\n", maxKey, pairs[i], pairs[i+1])
	}
	return b.String()
}

// RegisterAll registers all built-in slash commands.
func RegisterAll(r *Registry) {
	// Core
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

	// Engine & Session
	r.Register(compactCmd{})
	r.Register(contextCmd{})
	r.Register(summaryCmd{})
	r.Register(rewindCmd{})
	r.Register(continueCmd{})
	r.Register(turnsCmd{})
	r.Register(configCmd{})
	r.Register(exportCmd{})
	r.Register(shareCmd{})
	r.Register(copyCmd{})
	r.Register(tagCmd{})
	r.Register(resumeCmd{})
	r.Register(sessionCmd{})

	// Subsystem
	r.Register(memoryCmd{})
	r.Register(hooksCmd{})
	r.Register(skillsCmd{})
	r.Register(tasksCmd{})
	r.Register(agentsCmd{})
	r.Register(pluginCmd{})
	r.Register(reloadCmd{})
	r.Register(mcpCmd{})
	r.Register(bridgeCmd{})
	r.Register(loginCmd{})
	r.Register(feedbackCmd{})
	r.Register(onboardingCmd{})

	// Config & Mode
	r.Register(fastCmd{})
	r.Register(effortCmd{})
	r.Register(passesCmd{})
	r.Register(themeCmd{})
	r.Register(styleCmd{})
	r.Register(keybindCmd{})
	r.Register(vimCmd{})
	r.Register(voiceCmd{})
	r.Register(privacyCmd{})
	r.Register(initCmd{})

	// Git & Info
	r.Register(doctorCmd{})
	r.Register(diffCmd{})
	r.Register(branchCmd{})
	r.Register(commitCmd{})
	r.Register(issueCmd{})
	r.Register(prCommentsCmd{})
	r.Register(filesCmd{})
	r.Register(releaseNotesCmd{})
	r.Register(upgradeCmd{})
}
