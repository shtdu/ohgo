package coordinator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_ParseValidYAML(t *testing.T) {
	dir := t.TempDir()

	yamlContent := []byte(`name: researcher
description: Research agent
prompt: You are a research assistant.
model: claude-sonnet-4-20250514
tools:
  - web_search
  - file_read
max_turns: 10
`)
	err := os.WriteFile(filepath.Join(dir, "researcher.yaml"), yamlContent, 0o644)
	require.NoError(t, err)

	loader := NewLoader(dir)
	defs, err := loader.LoadAll(context.Background())
	require.NoError(t, err)

	require.Len(t, defs, 1)
	d := defs[0]
	assert.Equal(t, "researcher", d.Name)
	assert.Equal(t, "Research agent", d.Description)
	assert.Equal(t, "You are a research assistant.", d.Prompt)
	assert.Equal(t, "claude-sonnet-4-20250514", d.Model)
	assert.Equal(t, []string{"web_search", "file_read"}, d.Tools)
	assert.Equal(t, 10, d.MaxTurns)
}

func TestLoader_EmptyDirReturnsEmptySlice(t *testing.T) {
	dir := t.TempDir()

	loader := NewLoader(dir)
	defs, err := loader.LoadAll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, defs)
}

func TestLoader_InvalidYAMLSkipped(t *testing.T) {
	dir := t.TempDir()

	// Valid YAML
	validContent := []byte(`name: good-agent
description: A good agent
prompt: Hello
`)
	err := os.WriteFile(filepath.Join(dir, "good.yaml"), validContent, 0o644)
	require.NoError(t, err)

	// Invalid YAML
	invalidContent := []byte(`name: broken
description: [
invalid yaml here
`)
	err = os.WriteFile(filepath.Join(dir, "bad.yaml"), invalidContent, 0o644)
	require.NoError(t, err)

	// Non-YAML file should be ignored
	err = os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("not yaml"), 0o644)
	require.NoError(t, err)

	loader := NewLoader(dir)
	defs, err := loader.LoadAll(context.Background())
	require.NoError(t, err)

	require.Len(t, defs, 1)
	assert.Equal(t, "good-agent", defs[0].Name)
}

func TestLoader_MissingNameSkipped(t *testing.T) {
	dir := t.TempDir()

	content := []byte(`description: No name agent
prompt: Hello
`)
	err := os.WriteFile(filepath.Join(dir, "noname.yaml"), content, 0o644)
	require.NoError(t, err)

	loader := NewLoader(dir)
	defs, err := loader.LoadAll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, defs)
}

func TestLoader_NonexistentDir(t *testing.T) {
	loader := NewLoader("/nonexistent/path")
	defs, err := loader.LoadAll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, defs)
}

func TestLoader_MultipleDirs(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	yaml1 := []byte(`name: agent-a
description: Agent A
prompt: Prompt A
`)
	yaml2 := []byte(`name: agent-b
description: Agent B
prompt: Prompt B
`)
	err := os.WriteFile(filepath.Join(dir1, "a.yaml"), yaml1, 0o644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir2, "b.yaml"), yaml2, 0o644)
	require.NoError(t, err)

	loader := NewLoader(dir1, dir2)
	defs, err := loader.LoadAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, defs, 2)
}

func TestLoader_ContextCancellation(t *testing.T) {
	dir := t.TempDir()

	yamlContent := []byte(`name: agent
description: test
prompt: test
`)
	err := os.WriteFile(filepath.Join(dir, "agent.yaml"), yamlContent, 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	loader := NewLoader(dir)
	_, err = loader.LoadAll(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}
