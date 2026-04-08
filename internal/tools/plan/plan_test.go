package plan

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/permissions"
	"github.com/shtdu/ohgo/internal/tools"
)

func newTestChecker() *permissions.DefaultChecker {
	return &permissions.DefaultChecker{ /* mode defaults to zero value */ }
}

func TestEnterPlanModeTool_Name(t *testing.T) {
	tool := EnterPlanModeTool{}
	assert.Equal(t, "enter_plan_mode", tool.Name())
}

func TestExitPlanModeTool_Name(t *testing.T) {
	tool := ExitPlanModeTool{}
	assert.Equal(t, "exit_plan_mode", tool.Name())
}

func TestEnterPlanModeTool_Description(t *testing.T) {
	tool := EnterPlanModeTool{}
	assert.NotEmpty(t, tool.Description())
}

func TestExitPlanModeTool_Description(t *testing.T) {
	tool := ExitPlanModeTool{}
	assert.NotEmpty(t, tool.Description())
}

func TestEnterPlanModeTool_InputSchema(t *testing.T) {
	tool := EnterPlanModeTool{}
	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
	props := schema["properties"].(map[string]any)
	assert.Empty(t, props)
}

func TestExitPlanModeTool_InputSchema(t *testing.T) {
	tool := ExitPlanModeTool{}
	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
	props := schema["properties"].(map[string]any)
	assert.Empty(t, props)
}

func TestEnterPlanModeTool_SetsModeToPlan(t *testing.T) {
	checker := newTestChecker()
	tool := EnterPlanModeTool{Checker: checker}
	args, _ := json.Marshal(map[string]any{})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "plan mode")
	assert.Equal(t, permissions.ModePlan, checker.Mode())
}

func TestExitPlanModeTool_SetsModeToDefault(t *testing.T) {
	checker := newTestChecker()
	checker.SetMode(permissions.ModePlan)

	tool := ExitPlanModeTool{Checker: checker}
	args, _ := json.Marshal(map[string]any{})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "default")
	assert.Equal(t, permissions.ModeDefault, checker.Mode())
}

func TestEnterPlanModeThenExit(t *testing.T) {
	checker := newTestChecker()

	enter := EnterPlanModeTool{Checker: checker}
	exit := ExitPlanModeTool{Checker: checker}
	args, _ := json.Marshal(map[string]any{})

	// Enter plan mode.
	result, err := enter.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, permissions.ModePlan, checker.Mode())

	// Exit plan mode.
	result, err = exit.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, permissions.ModeDefault, checker.Mode())
}

func TestEnterPlanModeTool_NilChecker(t *testing.T) {
	tool := EnterPlanModeTool{Checker: nil}
	args, _ := json.Marshal(map[string]any{})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not available")
}

func TestExitPlanModeTool_NilChecker(t *testing.T) {
	tool := ExitPlanModeTool{Checker: nil}
	args, _ := json.Marshal(map[string]any{})

	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not available")
}

func TestEnterPlanModeTool_InvalidJSON(t *testing.T) {
	checker := newTestChecker()
	// Record the mode before the call to verify it does not change.
	modeBefore := checker.Mode()
	tool := EnterPlanModeTool{Checker: checker}

	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
	// Mode should not have changed.
	assert.Equal(t, modeBefore, checker.Mode())
}

func TestExitPlanModeTool_InvalidJSON(t *testing.T) {
	checker := newTestChecker()
	checker.SetMode(permissions.ModePlan)
	tool := ExitPlanModeTool{Checker: checker}

	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
	// Mode should not have changed.
	assert.Equal(t, permissions.ModePlan, checker.Mode())
}

// Verify both tools satisfy the interface.
var _ tools.Tool = EnterPlanModeTool{}
var _ tools.Tool = ExitPlanModeTool{}
