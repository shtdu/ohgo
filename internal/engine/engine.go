// Package engine implements the core agent loop: query -> stream -> tool_use -> loop.
package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/shtdu/ohgo/internal/api"
	"github.com/shtdu/ohgo/internal/hooks"
	"github.com/shtdu/ohgo/internal/permissions"
	"github.com/shtdu/ohgo/internal/tools"
)

// Options configures the engine.
type Options struct {
	Model      string
	MaxTokens  int
	MaxTurns   int
	System     string
	Permission permissions.Checker
	ToolReg    *tools.Registry
	Hooks      *hooks.Executor
	APIClient  api.Client
	EventCh    chan<- EngineEvent
}

// Engine drives the core agent loop.
type Engine struct {
	opts        Options
	messages    []api.Message
	costTracker *CostTracker
}

// New creates a new Engine with the given options.
func New(opts Options) *Engine {
	if opts.MaxTurns == 0 {
		opts.MaxTurns = 200
	}
	return &Engine{
		opts:        opts,
		costTracker: NewCostTracker(),
	}
}

// Query sends a user prompt through the agent loop and streams the response.
func (e *Engine) Query(ctx context.Context, prompt string) error {
	// Append user message
	e.messages = append(e.messages, api.NewUserTextMessage(prompt))

	for turn := 0; turn < e.opts.MaxTurns; turn++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Build API request
		apiTools := e.buildToolDefs()
		opts := api.StreamOptions{
			Model:     e.opts.Model,
			MaxTokens: e.opts.MaxTokens,
			System:    e.opts.System,
			Messages:  e.messages,
			Tools:     apiTools,
		}

		// Stream from API
		eventCh, err := e.opts.APIClient.Stream(ctx, opts)
		if err != nil {
			return fmt.Errorf("api stream: %w", err)
		}

		// Process events
		var contentBlocks []api.ContentBlock
		var usage api.UsageSnapshot
		for event := range eventCh {
			switch data := event.Data.(type) {
			case string:
				if event.Type == "text_delta" {
					e.emit(EngineEvent{Type: EventTextDelta, Data: AssistantTextDelta{Text: data}})
				} else if event.Type == "error" {
					e.emit(EngineEvent{Type: EventError, Data: ErrorEvent{Message: data, Recoverable: false}})
					return fmt.Errorf("api error: %s", data)
				}
			case api.Message:
				if event.Type == "message_complete" {
					contentBlocks = data.Content
				}
			case api.UsageSnapshot:
				if event.Type == "usage" {
					usage = data
					e.costTracker.Add(data)
				}
			}
		}

		// Build assistant message from collected blocks
		if contentBlocks == nil {
			contentBlocks = []api.ContentBlock{}
		}
		assistantMsg := api.NewAssistantMessage(contentBlocks)
		e.messages = append(e.messages, assistantMsg)
		e.costTracker.IncrementTurns()

		e.emit(EngineEvent{Type: EventTurnComplete, Data: AssistantTurnComplete{
			InputTokens:  usage.InputTokens,
			OutputTokens: usage.OutputTokens,
		}})

		// Check for tool_use
		toolCalls := ExtractToolCalls(assistantMsg)
		if len(toolCalls) == 0 {
			return nil
		}

		// Execute tools
		results := make([]ToolCallResult, 0, len(toolCalls))
		for _, call := range toolCalls {
			e.emit(EngineEvent{Type: EventToolStarted, Data: ToolExecutionStarted{
				ToolName:  call.Name,
				ToolInput: string(call.Input),
			}})

			output, isErr := e.executeTool(ctx, call)
			results = append(results, ToolCallResult{
				ToolUseID: call.ID,
				Content:   output,
				IsError:   isErr,
			})

			e.emit(EngineEvent{Type: EventToolCompleted, Data: ToolExecutionCompleted{
				ToolName: call.Name,
				Output:   output,
				IsError:  isErr,
			}})
		}

		// Append tool results
		e.messages = append(e.messages, BuildToolResultMessage(results))
	}

	return fmt.Errorf("max turns (%d) exceeded", e.opts.MaxTurns)
}

// executeTool runs a single tool with permission checks and hooks.
func (e *Engine) executeTool(ctx context.Context, call api.ToolCall) (string, bool) {
	tool := e.opts.ToolReg.Get(call.Name)
	if tool == nil {
		return fmt.Sprintf("unknown tool: %s", call.Name), true
	}

	// Run pre-hooks
	if e.opts.Hooks != nil {
		hookResp, err := e.opts.Hooks.RunPre(ctx, hooks.HookContext{
			Event:    hooks.PreToolUse,
			ToolName: call.Name,
		})
		if err != nil {
			return fmt.Sprintf("hook error: %v", err), true
		}
		if hookResp.Block {
			return fmt.Sprintf("blocked by hook: %s", hookResp.Reason), true
		}
	}

	// Check permissions
	if e.opts.Permission != nil {
		decision, err := e.opts.Permission.Check(ctx, permissions.Check{ToolName: call.Name})
		if err != nil {
			return fmt.Sprintf("permission check error: %v", err), true
		}
		if decision == permissions.Deny {
			return fmt.Sprintf("tool %s denied by permissions", call.Name), true
		}
	}

	// Execute
	result, err := tool.Execute(ctx, call.Input)
	if err != nil {
		return fmt.Sprintf("tool error: %v", err), true
	}

	// Run post-hooks
	if e.opts.Hooks != nil {
		_ = e.opts.Hooks.RunPost(ctx, hooks.HookContext{
			Event:    hooks.PostToolUse,
			ToolName: call.Name,
			Result:   result,
		})
	}

	return result.Content, result.IsError
}

// buildToolDefs converts registered tools to API tool definitions.
func (e *Engine) buildToolDefs() []api.ToolDef {
	if e.opts.ToolReg == nil {
		return nil
	}
	toolList := e.opts.ToolReg.List()
	defs := make([]api.ToolDef, 0, len(toolList))
	for _, t := range toolList {
		defs = append(defs, api.ToolDef{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.InputSchema(),
		})
	}
	return defs
}

// emit sends an event to the event channel if configured.
func (e *Engine) emit(event EngineEvent) {
	if e.opts.EventCh != nil {
		select {
		case e.opts.EventCh <- event:
		default:
			log.Printf("engine event channel full, dropping event type %d", event.Type)
		}
	}
}

// Clear resets conversation history and cost tracking.
func (e *Engine) Clear() {
	e.messages = nil
	e.costTracker.Reset()
}

// Messages returns a copy of the conversation history.
func (e *Engine) Messages() []api.Message {
	out := make([]api.Message, len(e.messages))
	copy(out, e.messages)
	return out
}

// TotalUsage returns the aggregated token usage.
func (e *Engine) TotalUsage() api.UsageSnapshot {
	return e.costTracker.Total()
}

// SetModel updates the model.
func (e *Engine) SetModel(model string) {
	e.opts.Model = model
}

// SetSystemPrompt updates the system prompt.
func (e *Engine) SetSystemPrompt(prompt string) {
	e.opts.System = prompt
}

// SetAPIClient updates the API client.
func (e *Engine) SetAPIClient(client api.Client) {
	e.opts.APIClient = client
}

// SetMaxTurns updates the max turn count.
func (e *Engine) SetMaxTurns(max int) {
	e.opts.MaxTurns = max
}

// LoadMessages replaces the conversation history.
func (e *Engine) LoadMessages(msgs []api.Message) {
	e.messages = make([]api.Message, len(msgs))
	copy(e.messages, msgs)
}

// Ensure api.Message, api.UsageSnapshot satisfy the type assertions used in Query
var _ = json.RawMessage{}
