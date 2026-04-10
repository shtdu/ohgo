// Package api defines the Client interface for communicating with LLM providers.
//
// This file implements the Anthropic provider using the official anthropic-sdk-go.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
)

// AnthropicClient implements the Client interface using the official Anthropic Go SDK.
type AnthropicClient struct {
	client anthropic.Client
}

// AnthropicOption configures an AnthropicClient.
type AnthropicOption func(*sdkConfig)

type sdkConfig struct {
	apiKey     string
	baseURL    string
	maxRetries int
}

// WithAPIKey sets the API key.
func WithAPIKey(key string) AnthropicOption {
	return func(c *sdkConfig) { c.apiKey = key }
}

// WithBaseURL sets a custom API base URL.
func WithBaseURL(url string) AnthropicOption {
	return func(c *sdkConfig) { c.baseURL = url }
}

// WithMaxRetries sets the maximum number of retries.
func WithMaxRetries(n int) AnthropicOption {
	return func(c *sdkConfig) { c.maxRetries = n }
}

// NewAnthropicClient creates a new Anthropic API client backed by the official SDK.
func NewAnthropicClient(opts ...AnthropicOption) *AnthropicClient {
	cfg := &sdkConfig{
		maxRetries: maxRetries,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	sdkOpts := []option.RequestOption{
		option.WithAPIKey(cfg.apiKey),
	}
	if cfg.baseURL != "" {
		// Strip /v1/messages suffix for backward compat with the old baseURL format.
		sdkOpts = append(sdkOpts, option.WithBaseURL(normalizeBaseURL(cfg.baseURL)))
	}
	if cfg.maxRetries > 0 {
		sdkOpts = append(sdkOpts, option.WithMaxRetries(cfg.maxRetries))
	}

	client := anthropic.NewClient(sdkOpts...)
	return &AnthropicClient{
		client: client,
	}
}

func normalizeBaseURL(url string) string {
	return strings.TrimSuffix(url, "/v1/messages")
}

// Stream sends a request to the Anthropic API and returns a channel of events.
func (c *AnthropicClient) Stream(ctx context.Context, opts StreamOptions) (<-chan StreamEvent, error) {
	params := c.buildParams(opts)
	stream := c.client.Messages.NewStreaming(ctx, params)

	ch := make(chan StreamEvent, 64)
	go func() {
		defer close(ch)
		c.processStream(stream, ch)
	}()

	return ch, nil
}

func (c *AnthropicClient) buildParams(opts StreamOptions) anthropic.MessageNewParams {
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(opts.Model),
		MaxTokens: int64(opts.MaxTokens),
		Messages:  convertMessages(opts.Messages),
	}

	if opts.System != "" {
		params.System = []anthropic.TextBlockParam{{Text: opts.System}}
	}

	if len(opts.Tools) > 0 {
		params.Tools = convertTools(opts.Tools)
	}

	return params
}

func (c *AnthropicClient) processStream(stream *ssestream.Stream[anthropic.MessageStreamEventUnion], ch chan<- StreamEvent) {
	var msg anthropic.Message

	for stream.Next() {
		event := stream.Current()
		if err := msg.Accumulate(event); err != nil {
			ch <- StreamEvent{Type: "error", Data: err.Error()}
			return
		}

		switch ev := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch delta := ev.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				ch <- StreamEvent{Type: "text_delta", Data: delta.Text}
			}
		case anthropic.MessageStopEvent:
			apiMsg := sdkToMessage(msg)
			ch <- StreamEvent{Type: "message_complete", Data: apiMsg}
			ch <- StreamEvent{Type: "usage", Data: sdkToUsage(msg.Usage)}
			slog.Debug("anthropic usage",
				"input", msg.Usage.InputTokens,
				"output", msg.Usage.OutputTokens,
				"cache_read", msg.Usage.CacheReadInputTokens,
				"cache_created", msg.Usage.CacheCreationInputTokens)
		}
	}

	if err := stream.Err(); err != nil {
		ch <- StreamEvent{Type: "error", Data: sdkError(err).Error()}
	}
}

// convertMessages converts internal Messages to SDK MessageParams.
func convertMessages(msgs []Message) []anthropic.MessageParam {
	result := make([]anthropic.MessageParam, 0, len(msgs))
	for _, msg := range msgs {
		blocks := make([]anthropic.ContentBlockParamUnion, 0, len(msg.Content))
		for _, block := range msg.Content {
			switch block.Type {
			case "text":
				blocks = append(blocks, anthropic.NewTextBlock(block.Text))
			case "tool_use":
				blocks = append(blocks, anthropic.NewToolUseBlock(block.ID, json.RawMessage(block.Input), block.Name))
			case "tool_result":
				blocks = append(blocks, anthropic.NewToolResultBlock(block.ToolUseID, block.Content, block.IsError))
			}
		}
		switch msg.Role {
		case "user":
			result = append(result, anthropic.NewUserMessage(blocks...))
		case "assistant":
			result = append(result, anthropic.NewAssistantMessage(blocks...))
		}
	}
	return result
}

// convertTools converts internal ToolDefs to SDK ToolUnionParams.
func convertTools(tools []ToolDef) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, 0, len(tools))
	for _, t := range tools {
		result = append(result, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        t.Name,
				Description: param.NewOpt(t.Description),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: t.InputSchema,
				},
			},
		})
	}
	return result
}

// sdkToMessage converts an accumulated SDK Message to an internal Message.
func sdkToMessage(msg anthropic.Message) Message {
	blocks := make([]ContentBlock, 0, len(msg.Content))
	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			blocks = append(blocks, ContentBlock{Type: "text", Text: block.Text})
		case "tool_use":
			blocks = append(blocks, ContentBlock{
				Type:  "tool_use",
				ID:    block.ID,
				Name:  block.Name,
				Input: block.Input,
			})
		}
	}
	return NewAssistantMessage(blocks)
}

// sdkToUsage converts SDK Usage to internal UsageSnapshot.
func sdkToUsage(u anthropic.Usage) UsageSnapshot {
	return UsageSnapshot{
		InputTokens:              int(u.InputTokens),
		OutputTokens:             int(u.OutputTokens),
		CacheReadInputTokens:     int(u.CacheReadInputTokens),
		CacheCreationInputTokens: int(u.CacheCreationInputTokens),
	}
}

// sdkError maps an SDK error to an internal API error.
func sdkError(err error) error {
	if err == nil {
		return nil
	}

	// The SDK's apierror.Error is aliased as anthropic.Error and carries StatusCode.
	var sdkErr *anthropic.Error
	if errors.As(err, &sdkErr) {
		return TranslateAPIError(sdkErr.StatusCode, sdkErr.Error())
	}

	return &APIError{StatusCode: 0, Message: err.Error(), Retryable: false}
}
