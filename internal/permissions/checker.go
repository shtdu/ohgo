// Package permissions defines the permission checking system for tool execution.
package permissions

import (
	"context"
	"path/filepath"

	"github.com/shtdu/ohgo/internal/config"
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

// String returns a human-readable decision.
func (d Decision) String() string {
	switch d {
	case Allow:
		return "allow"
	case Deny:
		return "deny"
	case Ask:
		return "ask"
	default:
		return "unknown"
	}
}

// Check represents a permission check request.
type Check struct {
	ToolName  string
	Args      map[string]any
	FilePath  string // extracted from Args if tool operates on files
	Command   string // extracted from Args if tool runs commands
	IsReadOnly bool  // set by the caller using ClassifyTool
}

// Checker is the interface for permission decisions on tool execution.
type Checker interface {
	// Check evaluates whether a tool invocation is permitted.
	Check(ctx context.Context, check Check) (Decision, error)
}

// PathRule is a glob-based path permission rule.
type PathRule struct {
	Pattern string
	Allow   bool
}

// DefaultChecker evaluates tool permissions based on mode and rules.
type DefaultChecker struct {
	mode           Mode
	allowedTools   map[string]bool
	deniedTools    map[string]bool
	pathRules      []PathRule
	deniedCommands []string
}

// NewDefaultChecker creates a checker from PermissionSettings.
func NewDefaultChecker(settings config.PermissionSettings) *DefaultChecker {
	allowed := make(map[string]bool, len(settings.AllowedTools))
	for _, t := range settings.AllowedTools {
		allowed[t] = true
	}
	denied := make(map[string]bool, len(settings.DeniedTools))
	for _, t := range settings.DeniedTools {
		denied[t] = true
	}
	rules := make([]PathRule, len(settings.PathRules))
	for i, r := range settings.PathRules {
		rules[i] = PathRule{Pattern: r.Pattern, Allow: r.Allow}
	}

	return &DefaultChecker{
		mode:           ParseMode(settings.Mode),
		allowedTools:   allowed,
		deniedTools:    denied,
		pathRules:      rules,
		deniedCommands: settings.DeniedCommands,
	}
}

// SetMode updates the permission mode.
func (d *DefaultChecker) SetMode(mode Mode) {
	d.mode = mode
}

// Mode returns the current permission mode.
func (d *DefaultChecker) Mode() Mode {
	return d.mode
}

// Check evaluates tool permissions based on mode + rules.
// Evaluation order:
// 1. Explicit deny list -> Deny
// 2. Explicit allow list -> Allow
// 3. Path rules (filepath.Match) -> Deny if matched and !Allow
// 4. Command deny patterns (filepath.Match) -> Deny
// 5. Mode: Auto->Allow, read-only->Allow, Plan->Deny write, Default->Ask for write
func (d *DefaultChecker) Check(ctx context.Context, check Check) (Decision, error) {
	// 1. Explicit deny
	if d.deniedTools[check.ToolName] {
		return Deny, nil
	}

	// 2. Explicit allow
	if d.allowedTools[check.ToolName] {
		return Allow, nil
	}

	// 3. Path rules
	if check.FilePath != "" {
		for _, rule := range d.pathRules {
			matched, err := filepath.Match(rule.Pattern, check.FilePath)
			if err == nil && matched {
				if rule.Allow {
					return Allow, nil
				}
				return Deny, nil
			}
		}
	}

	// 4. Command deny patterns
	if check.Command != "" {
		for _, pattern := range d.deniedCommands {
			matched, err := filepath.Match(pattern, check.Command)
			if err == nil && matched {
				return Deny, nil
			}
		}
	}

	// 5. Mode-based
	switch d.mode {
	case ModeAuto:
		return Allow, nil
	default:
		if check.IsReadOnly {
			return Allow, nil
		}
		if d.mode == ModePlan {
			return Deny, nil
		}
		return Ask, nil
	}
}
