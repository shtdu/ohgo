package hooks

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

// HookEvent identifies when a hook fires.
type HookEvent string

const (
	HookEventPreToolUse  HookEvent = "pre_tool_use"
	HookEventPostToolUse HookEvent = "post_tool_use"
)

// HookType identifies the kind of hook implementation.
type HookType string

const (
	HookTypeCommand HookType = "command"
	HookTypeHTTP    HookType = "http"
	HookTypePrompt  HookType = "prompt"
	HookTypeAgent   HookType = "agent"
)

// HookDefinition is the JSON-deserialized form of a hook entry.
type HookDefinition struct {
	Event          HookEvent         `json:"event"`
	Type           HookType          `json:"type"`
	Matcher        string            `json:"matcher,omitempty"`
	BlockOnFailure bool              `json:"block_on_failure,omitempty"`
	// Command hook
	Command        string            `json:"command,omitempty"`
	TimeoutSeconds int               `json:"timeout_seconds,omitempty"`
	// HTTP hook
	URL            string            `json:"url,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	// Prompt/Agent hook
	Prompt         string            `json:"prompt,omitempty"`
	Model          string            `json:"model,omitempty"`
}

// Validate checks that required fields are present for the hook type.
func (h HookDefinition) Validate() error {
	if h.Event == "" {
		return fmt.Errorf("hook event is required")
	}
	if h.Type == "" {
		return fmt.Errorf("hook type is required")
	}
	switch h.Type {
	case HookTypeCommand:
		if h.Command == "" {
			return fmt.Errorf("command hook requires 'command' field")
		}
	case HookTypeHTTP:
		if h.URL == "" {
			return fmt.Errorf("http hook requires 'url' field")
		}
	case HookTypePrompt, HookTypeAgent:
		if h.Prompt == "" {
			return fmt.Errorf("%s hook requires 'prompt' field", h.Type)
		}
	default:
		return fmt.Errorf("unknown hook type: %s", h.Type)
	}
	return nil
}

// HookResult is the outcome of a single hook execution.
type HookResult struct {
	HookType HookType
	Success  bool
	Output   string
	Blocked  bool
	Reason   string
}

// AggregatedResult collects results from all hooks for one event.
type AggregatedResult struct {
	Results []HookResult
}

// Blocked returns true if any hook blocked continuation.
func (ar *AggregatedResult) Blocked() bool {
	for _, r := range ar.Results {
		if r.Blocked {
			return true
		}
	}
	return false
}

// Reason returns the first blocking reason.
func (ar *AggregatedResult) Reason() string {
	for _, r := range ar.Results {
		if r.Blocked {
			return r.Reason
		}
	}
	return ""
}

// MatchesHook checks if a hook's matcher pattern matches the given subject.
// Uses filepath.Match (fnmatch-style) semantics.
// Empty matcher matches everything.
func MatchesHook(matcher string, subject string) bool {
	if matcher == "" {
		return true
	}
	matched, err := filepath.Match(matcher, subject)
	return err == nil && matched
}

// HookManifest is the top-level structure for hooks.json files.
type HookManifest struct {
	Hooks []json.RawMessage `json:"hooks"`
}
