package skill

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/skills"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestSkillTool_NameAndSchema(t *testing.T) {
	reg := skills.NewRegistry()
	tool := SkillTool{SkillReg: reg}

	assert.Equal(t, "skill", tool.Name())
	assert.Contains(t, tool.Description(), "skill")

	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
	required := schema["required"].([]string)
	assert.Contains(t, required, "skill_name")
}

func TestSkillTool_ExistingSkill(t *testing.T) {
	reg := skills.NewRegistry()
	reg.Register(&skills.Skill{
		Name:        "commit",
		Description: "Create a git commit",
		Content:     "# Commit Skill\n\n1. Stage changes\n2. Write message\n3. Commit",
	})

	tool := SkillTool{SkillReg: reg}
	args, _ := json.Marshal(map[string]any{
		"skill_name": "commit",
	})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "# Commit Skill")
	assert.Contains(t, result.Content, "Stage changes")
}

func TestSkillTool_MissingSkill(t *testing.T) {
	reg := skills.NewRegistry()
	tool := SkillTool{SkillReg: reg}

	args, _ := json.Marshal(map[string]any{
		"skill_name": "nonexistent",
	})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

func TestSkillTool_WithArgs(t *testing.T) {
	reg := skills.NewRegistry()
	reg.Register(&skills.Skill{
		Name:        "review-pr",
		Description: "Review a pull request",
		Content:     "# PR Review Skill\n\nReview the changes carefully.",
	})

	tool := SkillTool{SkillReg: reg}
	args, _ := json.Marshal(map[string]any{
		"skill_name": "review-pr",
		"args":       "PR #123",
	})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "# PR Review Skill")
	assert.Contains(t, result.Content, "PR #123")
	assert.Contains(t, result.Content, "Skill arguments:")
}

func TestSkillTool_EmptyArgs(t *testing.T) {
	reg := skills.NewRegistry()
	reg.Register(&skills.Skill{
		Name:    "test-skill",
		Content: "Just the content",
	})

	tool := SkillTool{SkillReg: reg}
	args, _ := json.Marshal(map[string]any{
		"skill_name": "test-skill",
		"args":       "",
	})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "Just the content", result.Content)
}

func TestSkillTool_InvalidJSON(t *testing.T) {
	reg := skills.NewRegistry()
	tool := SkillTool{SkillReg: reg}

	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestSkillTool_MissingSkillName(t *testing.T) {
	reg := skills.NewRegistry()
	tool := SkillTool{SkillReg: reg}

	args, _ := json.Marshal(map[string]any{
		"args": "some args",
	})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "skill_name is required")
}

func TestSkillTool_ContextCancel(t *testing.T) {
	reg := skills.NewRegistry()
	tool := SkillTool{SkillReg: reg}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	args, _ := json.Marshal(map[string]any{
		"skill_name": "anything",
	})

	_, err := tool.Execute(ctx, args)
	assert.Error(t, err)
}

// Verify the tool satisfies the interface at compile time.
var _ tools.Tool = SkillTool{}
