package plan

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shtdu/ohgo/internal/permissions"
	"github.com/shtdu/ohgo/internal/tools"
)

// ExitPlanModeTool switches the permission checker back to default mode.
type ExitPlanModeTool struct {
	Checker *permissions.DefaultChecker
}

func (ExitPlanModeTool) Name() string { return "exit_plan_mode" }

func (ExitPlanModeTool) Description() string {
	return "Exit plan mode and return to the default permission mode."
}

func (ExitPlanModeTool) InputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":          map[string]any{},
		"required":            []string{},
		"additionalProperties": false,
	}
}

func (t ExitPlanModeTool) Execute(_ context.Context, args json.RawMessage) (tools.Result, error) {
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

	t.Checker.SetMode(permissions.ModeDefault)
	return tools.Result{Content: "Exited plan mode. Returned to default permission mode."}, nil
}
