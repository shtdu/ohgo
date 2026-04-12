// Package hooks defines the PreToolUse/PostToolUse lifecycle hook system.
package hooks

import "context"

// HookRunner is the interface for running lifecycle hooks.
// Both the legacy Executor and DefinitionExecutor implement this.
//
// Contract:
//   - Pre-hooks run in registration order. The first hook to block wins; remaining hooks are skipped.
//   - Post-hooks all run regardless of individual errors. Errors are logged, not fatal.
//   - Hooks must complete within 30 seconds or context cancellation kills them.
type HookRunner interface {
	// RunPre executes pre-tool hooks. Returns true if execution should be blocked.
	RunPre(ctx context.Context, toolName string, args map[string]any) (blocked bool, reason string, err error)
	// RunPost executes post-tool hooks.
	RunPost(ctx context.Context, toolName string, args map[string]any, result any) error
}

// noopRunner is a HookRunner that does nothing.
type noopRunner struct{}

func (noopRunner) RunPre(_ context.Context, _ string, _ map[string]any) (bool, string, error) {
	return false, "", nil
}

func (noopRunner) RunPost(_ context.Context, _ string, _ map[string]any, _ any) error {
	return nil
}

// NoopRunner returns a HookRunner that does nothing.
func NoopRunner() HookRunner {
	return noopRunner{}
}
