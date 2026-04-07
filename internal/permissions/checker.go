// Package permissions defines the permission checking system for tool execution.
package permissions

import (
	"context"
)

// Decision represents the outcome of a permission check.
type Decision int

const (
	// Allow permits the action without prompting.
	Allow Decision = iota
	// Deny blocks the action.
	Deny
	// Ask prompts the user for approval.
	Ask
)

// Check represents a permission check request.
type Check struct {
	ToolName string
	Args     map[string]any
}

// Checker is the interface for permission decisions on tool execution.
type Checker interface {
	// Check evaluates whether a tool invocation is permitted.
	Check(ctx context.Context, check Check) (Decision, error)
}

// Mode represents a permission mode (e.g. "default", "plan", "autoedit").
type Mode string

const (
	ModeDefault Mode = "default"
	ModePlan    Mode = "plan"
	ModeAuto    Mode = "auto"
)

// DefaultChecker is a simple mode-based permission checker.
type DefaultChecker struct {
	mode Mode
}

// NewDefaultChecker creates a checker with the given permission mode.
func NewDefaultChecker(mode Mode) *DefaultChecker {
	return &DefaultChecker{mode: mode}
}

// Check evaluates tool permissions based on the configured mode.
func (d *DefaultChecker) Check(ctx context.Context, check Check) (Decision, error) {
	// TODO: implement mode-based permission logic
	return Ask, nil
}
