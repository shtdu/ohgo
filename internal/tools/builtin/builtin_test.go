package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestRegisterAll(t *testing.T) {
	r := tools.NewRegistry()
	RegisterAll(r)

	expectedTools := []string{
		"read_file", "write_file", "edit_file", "bash",
		"glob", "grep", "web_fetch", "web_search", "lsp",
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

	// Verify count
	all := r.List()
	assert.Len(t, all, len(expectedTools))

	// Verify no duplicate names
	names := make(map[string]bool)
	for _, tool := range all {
		assert.False(t, names[tool.Name()], "duplicate tool: %s", tool.Name())
		names[tool.Name()] = true
	}
}
