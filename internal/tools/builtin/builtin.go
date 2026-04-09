// Package builtin registers all built-in tools.
package builtin

import (
	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/coordinator"
	"github.com/shtdu/ohgo/internal/mcp"
	"github.com/shtdu/ohgo/internal/permissions"
	"github.com/shtdu/ohgo/internal/plugins"
	"github.com/shtdu/ohgo/internal/skills"
	"github.com/shtdu/ohgo/internal/tasks"
	"github.com/shtdu/ohgo/internal/tools"
	"github.com/shtdu/ohgo/internal/tools/agent"
	"github.com/shtdu/ohgo/internal/tools/ask"
	"github.com/shtdu/ohgo/internal/tools/bash"
	"github.com/shtdu/ohgo/internal/tools/brief"
	toolconfig "github.com/shtdu/ohgo/internal/tools/config"
	toolcron "github.com/shtdu/ohgo/internal/tools/cron"
	"github.com/shtdu/ohgo/internal/tools/edit"
	"github.com/shtdu/ohgo/internal/tools/glob"
	"github.com/shtdu/ohgo/internal/tools/grep"
	"github.com/shtdu/ohgo/internal/tools/lsp"
	mcptool "github.com/shtdu/ohgo/internal/tools/mcp"
	"github.com/shtdu/ohgo/internal/tools/message"
	"github.com/shtdu/ohgo/internal/tools/notebook"
	"github.com/shtdu/ohgo/internal/tools/plan"
	"github.com/shtdu/ohgo/internal/tools/read"
	"github.com/shtdu/ohgo/internal/tools/remote"
	"github.com/shtdu/ohgo/internal/tools/search"
	"github.com/shtdu/ohgo/internal/tools/skill"
	"github.com/shtdu/ohgo/internal/tools/sleep"
	teamtool "github.com/shtdu/ohgo/internal/tools/team"
	tooltask "github.com/shtdu/ohgo/internal/tools/task"
	"github.com/shtdu/ohgo/internal/tools/todo"
	"github.com/shtdu/ohgo/internal/tools/webfetch"
	"github.com/shtdu/ohgo/internal/tools/websearch"
	"github.com/shtdu/ohgo/internal/tools/worktree"
	"github.com/shtdu/ohgo/internal/tools/write"
)

// ToolDeps carries shared services needed by stateful tools.
// Fields may be nil; tools that require a nil dependency are not registered.
type ToolDeps struct {
	Checker   *permissions.DefaultChecker
	Settings  *config.Settings
	Registry  *tools.Registry
	CronMgr   *toolcron.Manager
	SkillReg  *skills.Registry
	TaskMgr   *tasks.Manager
	PluginMgr *plugins.Manager
	MCPMgr    *mcp.Manager
	Coord     *coordinator.Coordinator
	AskPrompter ask.Prompter
	MsgEmitter  func(msg message.Message) error
}

// RegisterAll registers all built-in tools into the registry.
// Stateless tools are always registered. Stateful tools are registered
// only if their required dependencies are non-nil.
func RegisterAll(r *tools.Registry, deps ToolDeps) {
	// Phase 3: Essential tools (stateless)
	r.Register(read.ReadTool{})
	r.Register(write.WriteTool{})
	r.Register(edit.EditTool{})
	r.Register(bash.BashTool{})
	r.Register(glob.GlobTool{})
	r.Register(grep.GrepTool{})
	r.Register(webfetch.FetchTool{})
	r.Register(websearch.SearchTool{})
	r.Register(lsp.LspTool{})

	// Phase 4: Batch 1 — Pure standalone
	r.Register(sleep.SleepTool{})
	r.Register(brief.BriefTool{})
	if deps.Registry != nil {
		r.Register(search.SearchTool{Registry: deps.Registry})
	}

	// Phase 4: Batch 2a — Config + Plan mode
	if deps.Settings != nil {
		r.Register(toolconfig.ConfigTool{Settings: deps.Settings})
	}
	if deps.Checker != nil {
		r.Register(plan.EnterPlanModeTool{Checker: deps.Checker})
		r.Register(plan.ExitPlanModeTool{Checker: deps.Checker})
	}

	// Phase 4: Batch 2b — Todo + Worktree
	r.Register(todo.TodoWriteTool{})
	r.Register(worktree.EnterWorktreeTool{})
	r.Register(worktree.ExitWorktreeTool{})

	// Phase 4: Batch 3 — Notebook edit
	r.Register(notebook.NotebookEditTool{})

	// Phase 4: Batch 4 — Cron tools
	if deps.CronMgr != nil {
		r.Register(toolcron.CreateTool{Mgr: deps.CronMgr})
		r.Register(toolcron.DeleteTool{Mgr: deps.CronMgr})
		r.Register(toolcron.ListTool{Mgr: deps.CronMgr})
		r.Register(toolcron.ToggleTool{Mgr: deps.CronMgr})
	}

	// Phase 4: Batch 5 — Task tools
	if deps.TaskMgr != nil {
		r.Register(tooltask.CreateTool{Mgr: deps.TaskMgr})
		r.Register(tooltask.GetTool{Mgr: deps.TaskMgr})
		r.Register(tooltask.ListTool{Mgr: deps.TaskMgr})
		r.Register(tooltask.OutputTool{Mgr: deps.TaskMgr})
		r.Register(tooltask.StopTool{Mgr: deps.TaskMgr})
		r.Register(tooltask.UpdateTool{Mgr: deps.TaskMgr})
	}

	// Phase 4: Batch 6 — Skill tool
	if deps.SkillReg != nil {
		r.Register(skill.SkillTool{SkillReg: deps.SkillReg})
	}

	// Phase 4: Batch 7 — Remote trigger (stateless)
	r.Register(remote.RemoteTriggerTool{})

	// Phase 4/6: MCP tools
	if deps.MCPMgr != nil {
		r.Register(mcptool.CallTool{Mgr: deps.MCPMgr})
		r.Register(mcptool.ListResources{Mgr: deps.MCPMgr})
		r.Register(mcptool.ReadResource{Mgr: deps.MCPMgr})
		r.Register(mcptool.Auth{Mgr: deps.MCPMgr})
	}

	// Phase 4/6: Coordinator tools
	if deps.Coord != nil {
		r.Register(agent.SpawnTool{Coord: deps.Coord})
		r.Register(teamtool.CreateTool{Coord: deps.Coord})
		r.Register(teamtool.DeleteTool{Coord: deps.Coord})
	}

	// Phase 4: Ask user question tool
	if deps.AskPrompter != nil {
		r.Register(ask.AskTool{Prompter: deps.AskPrompter})
	}

	// Phase 4: Send message tool
	if deps.MsgEmitter != nil {
		r.Register(message.SendTool{Emitter: deps.MsgEmitter})
	}
}
