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
	switch d.mode {
	case ModeAuto:
		return Allow, nil
	case ModePlan:
		// In plan mode, deny write tools
		writeTools := map[string]bool{
			"write_file": true, "edit_file": true, "bash": true,
		}
		if writeTools[check.ToolName] {
			return Deny, nil
		}
		return Allow, nil
	default:
		// Default mode: ask for write tools, allow read tools
		readTools := map[string]bool{
			"read_file": true, "glob": true, "grep": true,
		}
		if readTools[check.ToolName] {
			return Allow, nil
		}
		return Ask, nil
	}
}
