// Package hooks defines the PreToolUse/PostToolUse lifecycle hook system.
package hooks

import (
	"context"
)

// Event represents a hook event type.
type Event int

const (
	// PreToolUse fires before a tool is executed.
	PreToolUse Event = iota
	// PostToolUse fires after a tool has been executed.
	PostToolUse
)

// HookContext provides context for a hook invocation.
type HookContext struct {
	Event    Event
	ToolName string
	Args     map[string]any
	Result   any // only populated for PostToolUse
}

// HookResponse is the result of a hook execution.
type HookResponse struct {
	// Block prevents the tool from executing (PreToolUse only).
	Block bool
	// Reason explains why the hook blocked execution.
	Reason string
	// ModifiedArgs allows a PreToolUse hook to alter tool arguments.
	ModifiedArgs map[string]any
}

// Hook is a callback that runs before or after tool execution.
type Hook func(ctx context.Context, hctx HookContext) (HookResponse, error)

// Executor manages registered hooks and runs them in order.
type Executor struct {
	preHooks  []Hook
	postHooks []Hook
}

// NewExecutor creates a new hook executor.
func NewExecutor() *Executor {
	return &Executor{}
}

// RegisterPre adds a hook that runs before tool execution.
func (e *Executor) RegisterPre(h Hook) {
	e.preHooks = append(e.preHooks, h)
}

// RegisterPost adds a hook that runs after tool execution.
func (e *Executor) RegisterPost(h Hook) {
	e.postHooks = append(e.postHooks, h)
}

// RunPre executes all pre-tool hooks in order. Stops on first block.
func (e *Executor) RunPre(ctx context.Context, hctx HookContext) (HookResponse, error) {
	for _, h := range e.preHooks {
		resp, err := h(ctx, hctx)
		if err != nil {
			return resp, err
		}
		if resp.Block {
			return resp, nil
		}
	}
	return HookResponse{}, nil
}

// RunPost executes all post-tool hooks in order.
func (e *Executor) RunPost(ctx context.Context, hctx HookContext) error {
	for _, h := range e.postHooks {
		if _, err := h(ctx, hctx); err != nil {
			return err
		}
	}
	return nil
}
