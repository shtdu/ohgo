// Package skill implements the skill tool for invoking named skills from the registry.
package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shtdu/ohgo/internal/skills"
	"github.com/shtdu/ohgo/internal/tools"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type skillInput struct {
	SkillName string `json:"skill_name"`
	Args      string `json:"args"`
}

// SkillTool loads a named skill from the registry and returns its markdown content.
type SkillTool struct {
	SkillReg *skills.Registry
}

func (SkillTool) Name() string { return "skill" }

func (SkillTool) Description() string {
	return "Invokes a named skill by loading its content from the registry. The skill's markdown content is returned for the agent to follow."
}

func (SkillTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"skill_name": map[string]any{
				"type":        "string",
				"description": "Name of the skill to invoke",
			},
			"args": map[string]any{
				"type":        "string",
				"description": "Extra arguments to pass to the skill",
			},
		},
		"required":             []string{"skill_name"},
		"additionalProperties": false,
	}
}

func (t SkillTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	// Check context first
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	var input skillInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.SkillName == "" {
		return tools.Result{Content: "skill_name is required", IsError: true}, nil
	}

	// Look up skill in registry with case-insensitive fallbacks.
	s := t.SkillReg.Get(input.SkillName)
	if s == nil {
		s = t.SkillReg.Get(strings.ToLower(input.SkillName))
	}
	if s == nil {
		s = t.SkillReg.Get(cases.Title(language.English).String(input.SkillName))
	}
	if s == nil {
		return tools.Result{
			Content: fmt.Sprintf("skill %q not found", input.SkillName),
			IsError: true,
		}, nil
	}

	output := s.Content
	if input.Args != "" {
		output += "\n\n---\n**Skill arguments:** " + input.Args
	}

	return tools.Result{Content: output}, nil
}

// Verify the tool satisfies the interface.
var _ tools.Tool = SkillTool{}
