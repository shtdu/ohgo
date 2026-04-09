package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "https://api.anthropic.com/v1/messages"
	apiVersionHeader = "2023-06-01"
)

// AnthropicClient implements the Client interface for the Anthropic API.
type AnthropicClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

// AnthropicOption configures an AnthropicClient.
type AnthropicOption func(*AnthropicClient)

// WithAPIKey sets the API key.
func WithAPIKey(key string) AnthropicOption {
	return func(c *AnthropicClient) { c.apiKey = key }
}

// WithBaseURL sets a custom API base URL.
func WithBaseURL(url string) AnthropicOption {
	return func(c *AnthropicClient) { c.baseURL = url }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) AnthropicOption {
	return func(c *AnthropicClient) { c.httpClient = hc }
}

// WithMaxRetries sets the maximum number of retries.
func WithMaxRetries(n int) AnthropicOption {
	return func(c *AnthropicClient) { c.maxRetries = n }
}

// NewAnthropicClient creates a new Anthropic API client.
func NewAnthropicClient(opts ...AnthropicOption) *AnthropicClient {
	c := &AnthropicClient{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
		maxRetries: maxRetries,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Stream sends a request to the Anthropic API and returns a channel of events.
func (c *AnthropicClient) Stream(ctx context.Context, opts StreamOptions) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 64)

	go func() {
		defer close(ch)
		c.streamWithRetry(ctx, opts, ch)
	}()

	return ch, nil
}

func (c *AnthropicClient) streamWithRetry(ctx context.Context, opts StreamOptions, ch chan<- StreamEvent) {
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := c.streamOnce(ctx, opts, ch); err != nil {
			if !IsRetryable(err) || attempt >= c.maxRetries {
				ch <- StreamEvent{Type: "error", Data: err.Error()}
				return
			}
			delay := retryDelay(attempt)
			log.Printf("api retry attempt %d/%d after %s: %v", attempt+1, c.maxRetries, delay, err)
			select {
			case <-ctx.Done():
				ch <- StreamEvent{Type: "error", Data: ctx.Err().Error()}
				return
			case <-time.After(delay):
			}
			continue
		}
		return
	}
}

func (c *AnthropicClient) streamOnce(ctx context.Context, opts StreamOptions, ch chan<- StreamEvent) error {
	body, err := c.buildRequestBody(opts)
	if err != nil {
		return fmt.Errorf("build request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersionHeader)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIError{StatusCode: 0, Message: err.Error(), Retryable: true}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		apiErr := TranslateAPIError(resp.StatusCode, string(respBody))
		return apiErr
	}

	c.parseSSEStream(resp.Body, ch)
	return nil
}

func (c *AnthropicClient) buildRequestBody(opts StreamOptions) ([]byte, error) {
	req := map[string]any{
		"model":      opts.Model,
		"max_tokens": opts.MaxTokens,
		"stream":     true,
	}

	// Messages
	msgs := make([]any, 0, len(opts.Messages))
	for _, msg := range opts.Messages {
		msgs = append(msgs, msg)
	}
	req["messages"] = msgs

	// System prompt
	if opts.System != "" {
		req["system"] = opts.System
	}

	// Tools
	if len(opts.Tools) > 0 {
		tools := make([]map[string]any, 0, len(opts.Tools))
		for _, t := range opts.Tools {
			tools = append(tools, map[string]any{
				"name":         t.Name,
				"description":  t.Description,
				"input_schema": t.InputSchema,
			})
		}
		req["tools"] = tools
	}

	return json.Marshal(req)
}

func (c *AnthropicClient) parseSSEStream(reader io.Reader, ch chan<- StreamEvent) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var eventType string
	var contentBlocks []ContentBlock
	var toolInputParts []string // accumulates input_json_delta fragments
	var usage UsageSnapshot

	for scanner.Scan() {
		line := scanner.Text()

		if prefix, ok := strings.CutPrefix(line, "event: "); ok {
			eventType = prefix
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			if line == "" {
				eventType = ""
			}
			continue
		}

		data, _ := strings.CutPrefix(line, "data: ")

		switch eventType {
		case "message_start":
			contentBlocks = nil
			usage = UsageSnapshot{}
			var msg struct {
				Message struct {
					Usage struct {
						InputTokens  int `json:"input_tokens"`
						OutputTokens int `json:"output_tokens"`
					} `json:"usage"`
				} `json:"message"`
			}
			if json.Unmarshal([]byte(data), &msg) == nil {
				usage.InputTokens = msg.Message.Usage.InputTokens
			}

		case "content_block_start":
			toolInputParts = nil
			var block struct {
				ContentBlock ContentBlock `json:"content_block"`
			}
			if json.Unmarshal([]byte(data), &block) == nil {
				contentBlocks = append(contentBlocks, block.ContentBlock)
			}

		case "content_block_delta":
			var delta struct {
				Delta struct {
					Type        string `json:"type"`
					Text        string `json:"text"`
					PartialJSON string `json:"partial_json"`
				} `json:"delta"`
				Index int `json:"index"`
			}
			if json.Unmarshal([]byte(data), &delta) == nil {
				if delta.Delta.Type == "text_delta" && delta.Delta.Text != "" {
					ch <- StreamEvent{Type: "text_delta", Data: delta.Delta.Text}
				}
				if delta.Delta.Type == "input_json_delta" && delta.Delta.PartialJSON != "" {
					toolInputParts = append(toolInputParts, delta.Delta.PartialJSON)
				}
			}

		case "content_block_stop":
			if len(toolInputParts) > 0 && len(contentBlocks) > 0 {
				combined := strings.Join(toolInputParts, "")
				idx := len(contentBlocks) - 1
				contentBlocks[idx].Input = json.RawMessage(combined)
				toolInputParts = nil
			}

		case "message_delta":
			var msgDelta struct {
				Delta struct {
					StopReason string `json:"stop_reason"`
				} `json:"delta"`
				Usage struct {
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			}
			if json.Unmarshal([]byte(data), &msgDelta) == nil {
				usage.OutputTokens = msgDelta.Usage.OutputTokens
			}

		case "message_stop":
			// Emit the complete message
			msg := NewAssistantMessage(contentBlocks)
			ch <- StreamEvent{Type: "message_complete", Data: msg}
			ch <- StreamEvent{Type: "usage", Data: usage}
		}

		if line == "" {
			eventType = ""
		}
	}
}

