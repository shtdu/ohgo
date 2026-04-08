// Package builtin registers all built-in tools.
package builtin

import (
	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/permissions"
	"github.com/shtdu/ohgo/internal/tools"
	"github.com/shtdu/ohgo/internal/tools/bash"
	"github.com/shtdu/ohgo/internal/tools/brief"
	toolconfig "github.com/shtdu/ohgo/internal/tools/config"
	"github.com/shtdu/ohgo/internal/tools/edit"
	"github.com/shtdu/ohgo/internal/tools/glob"
	"github.com/shtdu/ohgo/internal/tools/grep"
	"github.com/shtdu/ohgo/internal/tools/lsp"
	"github.com/shtdu/ohgo/internal/tools/notebook"
	"github.com/shtdu/ohgo/internal/tools/plan"
	"github.com/shtdu/ohgo/internal/tools/read"
	"github.com/shtdu/ohgo/internal/tools/search"
	"github.com/shtdu/ohgo/internal/tools/sleep"
	"github.com/shtdu/ohgo/internal/tools/todo"
	toolcron "github.com/shtdu/ohgo/internal/tools/cron"
	"github.com/shtdu/ohgo/internal/tools/webfetch"
	"github.com/shtdu/ohgo/internal/tools/websearch"
	"github.com/shtdu/ohgo/internal/tools/worktree"
	"github.com/shtdu/ohgo/internal/tools/write"
)

// ToolDeps carries shared services needed by stateful tools.
// Fields may be nil; tools that require a nil dependency are not registered.
type ToolDeps struct {
	Checker  *permissions.DefaultChecker
	Settings *config.Settings
	Registry *tools.Registry
	CronMgr  *toolcron.Manager
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
}
