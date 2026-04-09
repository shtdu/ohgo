package commands

import (
	"context"
	"fmt"
	"strings"
)

type skillsCmd struct{}

var _ Command = skillsCmd{}

func (skillsCmd) Name() string      { return "skills" }
func (skillsCmd) ShortHelp() string { return "list loaded skills" }

func (skillsCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	if deps.Skills == nil {
		return Result{Output: "skills: no skills loaded"}, nil
	}

	skillList := deps.Skills.List()
	if len(skillList) == 0 {
		return Result{Output: "skills: no skills loaded"}, nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Loaded skills (%d):\n", len(skillList))
	for _, s := range skillList {
		fmt.Fprintf(&b, "  %s", s.Name)
		if s.Description != "" {
			fmt.Fprintf(&b, " - %s", s.Description)
		}
		fmt.Fprintln(&b)
	}
	return Result{Output: b.String()}, nil
}
