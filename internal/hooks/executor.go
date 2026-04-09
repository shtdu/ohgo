package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Payload is the data passed to hooks for a given event.
type Payload struct {
	ToolName string         `json:"tool_name,omitempty"`
	Args     map[string]any `json:"args,omitempty"`
	Result   any            `json:"result,omitempty"`
}

// DefinitionExecutor runs hooks from a Registry for lifecycle events.
// It implements the HookRunner interface for integration with the engine.
//
// Command hooks inject $ARGUMENTS directly into a shell command string.
// This is by design (hooks need access to tool args) but means hook commands
// run with the full privileges of the og process. Hook definitions should be
// treated as trusted configuration.
type DefinitionExecutor struct {
	mu       sync.RWMutex
	registry *Registry
	client   *http.Client
}

// NewDefinitionExecutor creates an executor backed by the given registry.
func NewDefinitionExecutor(registry *Registry) *DefinitionExecutor {
	if registry == nil {
		registry = NewRegistry()
	}
	return &DefinitionExecutor{
		registry: registry,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// UpdateRegistry replaces the active hook registry (for hot reload).
func (e *DefinitionExecutor) UpdateRegistry(registry *Registry) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.registry = registry
}

func (e *DefinitionExecutor) getRegistry() *Registry {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.registry
}

// RunPre implements HookRunner. Returns blocked=true if any hook blocks execution.
func (e *DefinitionExecutor) RunPre(ctx context.Context, toolName string, args map[string]any) (bool, string, error) {
	agg, err := e.ExecutePre(ctx, toolName, args)
	if err != nil {
		return false, "", err
	}
	if agg.Blocked() {
		return true, agg.Reason(), nil
	}
	return false, "", nil
}

// RunPost implements HookRunner.
func (e *DefinitionExecutor) RunPost(ctx context.Context, toolName string, args map[string]any, result any) error {
	_, err := e.ExecutePost(ctx, toolName, args, result)
	return err
}

// ExecutePre runs all pre-tool hooks. Returns aggregated result.
// If any hook blocks, execution stops and the aggregated result reflects it.
func (e *DefinitionExecutor) ExecutePre(ctx context.Context, toolName string, args map[string]any) (*AggregatedResult, error) {
	hooks := e.getRegistry().Get(HookEventPreToolUse)
	return e.executeHooks(ctx, hooks, toolName, args, nil)
}

// ExecutePost runs all post-tool hooks. Returns aggregated result.
func (e *DefinitionExecutor) ExecutePost(ctx context.Context, toolName string, args map[string]any, result any) (*AggregatedResult, error) {
	hooks := e.getRegistry().Get(HookEventPostToolUse)
	return e.executeHooks(ctx, hooks, toolName, args, result)
}

func (e *DefinitionExecutor) executeHooks(ctx context.Context, hookDefs []HookDefinition, toolName string, args map[string]any, result any) (*AggregatedResult, error) {
	var agg AggregatedResult
	for _, hook := range hookDefs {
		// Filter by matcher
		if !MatchesHook(hook.Matcher, toolName) {
			continue
		}

		select {
		case <-ctx.Done():
			return &agg, ctx.Err()
		default:
		}

		var hookResult HookResult
		switch hook.Type {
		case HookTypeCommand:
			hookResult = e.runCommandHook(ctx, hook, toolName, args)
		case HookTypeHTTP:
			hookResult = e.runHTTPHook(ctx, hook, toolName, args, result)
		case HookTypePrompt, HookTypeAgent:
			// Stub: prompt/agent hooks require an API client reference.
			// Full implementation deferred to a later phase.
			slog.Debug("hook type not yet implemented, skipping", "type", hook.Type, "matcher", hook.Matcher)
			hookResult = HookResult{HookType: hook.Type, Success: true}
		default:
			hookResult = HookResult{HookType: hook.Type, Success: true}
		}

		agg.Results = append(agg.Results, hookResult)
		if hookResult.Blocked {
			return &agg, nil
		}

		// Propagate context cancellation after hook execution.
		select {
		case <-ctx.Done():
			return &agg, ctx.Err()
		default:
		}
	}
	return &agg, nil
}

func (e *DefinitionExecutor) runCommandHook(ctx context.Context, hook HookDefinition, toolName string, args map[string]any) HookResult {
	timeout := 30 * time.Second
	if hook.TimeoutSeconds > 0 {
		timeout = time.Duration(hook.TimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Inject $ARGUMENTS into command
	argsJSON, _ := json.Marshal(args)
	command := strings.ReplaceAll(hook.Command, "$ARGUMENTS", string(argsJSON))

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Env = append(os.Environ(),
		"OPENHARNESS_HOOK_EVENT="+string(hook.Event),
		"OPENHARNESS_HOOK_TOOL="+toolName,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}

	result := HookResult{
		HookType: HookTypeCommand,
		Output:   output,
		Success:  err == nil,
	}

	if err != nil {
		result.Blocked = hook.BlockOnFailure
		result.Reason = fmt.Sprintf("command hook failed: %v", err)
	}

	return result
}

func (e *DefinitionExecutor) runHTTPHook(ctx context.Context, hook HookDefinition, toolName string, args map[string]any, result any) HookResult {
	payload := Payload{
		ToolName: toolName,
		Args:     args,
		Result:   result,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return HookResult{
			HookType: HookTypeHTTP,
			Success:  false,
			Blocked:  hook.BlockOnFailure,
			Reason:   fmt.Sprintf("marshal payload: %v", err),
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.URL, bytes.NewReader(body))
	if err != nil {
		return HookResult{
			HookType: HookTypeHTTP,
			Success:  false,
			Blocked:  hook.BlockOnFailure,
			Reason:   fmt.Sprintf("create request: %v", err),
		}
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range hook.Headers {
		req.Header.Set(k, v)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return HookResult{
			HookType: HookTypeHTTP,
			Success:  false,
			Blocked:  hook.BlockOnFailure,
			Reason:   fmt.Sprintf("http request: %v", err),
		}
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	return HookResult{
		HookType: HookTypeHTTP,
		Success:  success,
		Blocked:  !success && hook.BlockOnFailure,
		Reason:   fmt.Sprintf("HTTP %d", resp.StatusCode),
	}
}
