package plan

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/permissions"
	"github.com/shtdu/ohgo/internal/tools"
)

// EnterPlanModeTool switches the permission checker into plan mode.
type EnterPlanModeTool struct {
	Checker *permissions.DefaultChecker
}

func (EnterPlanModeTool) Name() string { return "enter_plan_mode" }

func (EnterPlanModeTool) Description() string {
	return "Enter plan mode, which restricts tool usage to read-only operations."
}

func (EnterPlanModeTool) InputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":          map[string]any{},
		"required":            []string{},
		"additionalProperties": false,
	}
}

func (t EnterPlanModeTool) Execute(_ context.Context, args json.RawMessage) (tools.Result, error) {
	// Validate that args is valid JSON (even though we expect empty object).
	if args != nil {
		var raw map[string]any
		if err := json.Unmarshal(args, &raw); err != nil {
			return tools.Result{
				Content: fmt.Sprintf("invalid arguments: %v", err),
				IsError: true,
			}, nil
		}
	}

	if t.Checker == nil {
		return tools.Result{
			Content: "permission checker is not available",
			IsError: true,
		}, nil
	}

	t.Checker.SetMode(permissions.ModePlan)
	return tools.Result{Content: "Entered plan mode. Tool usage is now restricted to read-only operations."}, nil
}
