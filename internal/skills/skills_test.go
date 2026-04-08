package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadByName_ValidSkill(t *testing.T) {
	dir := t.TempDir()
	content := `---
name: commit
description: Create a git commit
---
## Steps
1. Stage files
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "commit.md"), []byte(content), 0644))

	loader := NewLoader(dir)
	skill, err := loader.LoadByName(context.Background(), "commit")
	require.NoError(t, err)
	require.NotNil(t, skill)
	assert.Equal(t, "commit", skill.Name)
	assert.Equal(t, "Create a git commit", skill.Description)
	assert.Contains(t, skill.Content, "## Steps")
	assert.Equal(t, filepath.Join(dir, "commit.md"), skill.Path)
	assert.Equal(t, "file", skill.Source)
}

func TestLoadByName_MissingSkill(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir)
	skill, err := loader.LoadByName(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, skill)
}

func TestLoadByName_SearchesMultipleDirs(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir2, "found.md"), []byte("# Found\n\nDesc here.\n"), 0644))

	loader := NewLoader(dir1, dir2)
	skill, err := loader.LoadByName(context.Background(), "found")
	require.NoError(t, err)
	require.NotNil(t, skill)
	assert.Equal(t, "Found", skill.Name)
}

func TestLoadAll_MultipleDirsWithMultipleSkills(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir1, "alpha.md"), []byte("---\nname: alpha\ndescription: First\n---\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "beta.md"), []byte("---\nname: beta\ndescription: Second\n---\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "gamma.md"), []byte("---\nname: gamma\ndescription: Third\n---\n"), 0644))

	loader := NewLoader(dir1, dir2)
	skills, err := loader.LoadAll(context.Background())
	require.NoError(t, err)
	require.Len(t, skills, 3)
	assert.Equal(t, "alpha", skills[0].Name)
	assert.Equal(t, "beta", skills[1].Name)
	assert.Equal(t, "gamma", skills[2].Name)
}

func TestLoadAll_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	loader := NewLoader(dir)
	skills, err := loader.LoadAll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestLoadAll_NonexistentDir(t *testing.T) {
	loader := NewLoader("/nonexistent/path")
	skills, err := loader.LoadAll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestLoad_CancelledContext(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.md"), []byte("# Skill\n\nDesc.\n"), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	loader := NewLoader(dir)
	_, err := loader.Load(ctx, "skill")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestLoadAll_CancelledContext(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skill.md"), []byte("# Skill\n\nDesc.\n"), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	loader := NewLoader(dir)
	_, err := loader.LoadAll(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestLoadByName_WithMdExtension(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "commit.md"), []byte("---\nname: commit\ndescription: test\n---\n"), 0644))

	loader := NewLoader(dir)
	skill, err := loader.LoadByName(context.Background(), "commit.md")
	require.NoError(t, err)
	require.NotNil(t, skill)
	assert.Equal(t, "commit", skill.Name)
}
