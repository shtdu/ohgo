package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/permissions"
	toolcron "github.com/shtdu/ohgo/internal/tools/cron"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestRegisterAll_MinimalDeps(t *testing.T) {
	r := tools.NewRegistry()
	RegisterAll(r, ToolDeps{})

	// Stateless tools are always registered
	expectedTools := []string{
		"read_file", "write_file", "edit_file", "bash",
		"glob", "grep", "web_fetch", "web_search", "lsp",
		"sleep", "brief", "todo_write", "enter_worktree",
		"exit_worktree", "notebook_edit",
	}

	for _, name := range expectedTools {
		tool := r.Get(name)
		require.NotNil(t, tool, "missing tool: %s", name)
		assert.Equal(t, name, tool.Name())
		assert.NotEmpty(t, tool.Description())
		schema := tool.InputSchema()
		assert.NotNil(t, schema)
		assert.Equal(t, "object", schema["type"])
	}

	// Stateless tools only (no deps provided)
	all := r.List()
	assert.Len(t, all, len(expectedTools))

	// Verify no duplicate names
	names := make(map[string]bool)
	for _, tool := range all {
		assert.False(t, names[tool.Name()], "duplicate tool: %s", tool.Name())
		names[tool.Name()] = true
	}
}

func TestRegisterAll_WithDeps(t *testing.T) {
	r := tools.NewRegistry()
	checker := permissions.NewDefaultChecker(config.PermissionSettings{Mode: "default"})
	cronMgr := toolcron.NewManager()
	cfg := &config.Settings{Model: "test"}

	RegisterAll(r, ToolDeps{
		Checker:  checker,
		Settings: cfg,
		Registry: r,
		CronMgr:  cronMgr,
	})

	// All 23 tools should be registered
	expectedTools := []string{
		// Phase 3
		"read_file", "write_file", "edit_file", "bash",
		"glob", "grep", "web_fetch", "web_search", "lsp",
		// Phase 4: Batch 1
		"sleep", "brief", "tool_search",
		// Phase 4: Batch 2a
		"config", "enter_plan_mode", "exit_plan_mode",
		// Phase 4: Batch 2b
		"todo_write", "enter_worktree", "exit_worktree",
		// Phase 4: Batch 3
		"notebook_edit",
		// Phase 4: Batch 4
		"cron_create", "cron_delete", "cron_list", "cron_toggle",
	}

	for _, name := range expectedTools {
		tool := r.Get(name)
		require.NotNil(t, tool, "missing tool: %s", name)
		assert.Equal(t, name, tool.Name())
	}

	all := r.List()
	assert.Len(t, all, len(expectedTools))
}
