package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/shtdu/ohgo/internal/config"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1/chat/completions"

// OpenAIClient implements Client for OpenAI-compatible APIs.
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

// OpenAIClientOption configures an OpenAIClient.
type OpenAIClientOption func(*OpenAIClient)

// WithOpenAIAPIKey sets the API key.
func WithOpenAIAPIKey(key string) OpenAIClientOption {
	return func(c *OpenAIClient) { c.apiKey = key }
}

// WithOpenAIBaseURL sets a custom API base URL.
func WithOpenAIBaseURL(url string) OpenAIClientOption {
	return func(c *OpenAIClient) { c.baseURL = url }
}

// WithOpenAIHTTPClient sets a custom HTTP client.
func WithOpenAIHTTPClient(hc *http.Client) OpenAIClientOption {
	return func(c *OpenAIClient) { c.httpClient = hc }
}

// WithOpenAIMaxRetries sets the maximum number of retries.
func WithOpenAIMaxRetries(n int) OpenAIClientOption {
	return func(c *OpenAIClient) { c.maxRetries = n }
}

// NewOpenAIClient creates a new OpenAI-compatible API client.
func NewOpenAIClient(opts ...OpenAIClientOption) *OpenAIClient {
	c := &OpenAIClient{
		baseURL:    defaultOpenAIBaseURL,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
		maxRetries: maxRetries,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Stream sends a request to the OpenAI-compatible API and returns a channel of events.
func (c *OpenAIClient) Stream(ctx context.Context, opts StreamOptions) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 64)

	go func() {
		defer close(ch)
		c.streamWithRetry(ctx, opts, ch)
	}()

	return ch, nil
}

func (c *OpenAIClient) streamWithRetry(ctx context.Context, opts StreamOptions, ch chan<- StreamEvent) {
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := c.streamOnce(ctx, opts, ch); err != nil {
			if !IsRetryable(err) || attempt >= c.maxRetries {
				ch <- StreamEvent{Type: "error", Data: err.Error()}
				return
			}
			delay := retryDelay(attempt)
			log.Printf("openai retry attempt %d/%d after %s: %v", attempt+1, c.maxRetries, delay, err)
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

func (c *OpenAIClient) streamOnce(ctx context.Context, opts StreamOptions, ch chan<- StreamEvent) error {
	body, err := c.buildRequestBody(opts)
	if err != nil {
		return fmt.Errorf("build request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIError{StatusCode: 0, Message: err.Error(), Retryable: true}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return TranslateAPIError(resp.StatusCode, string(respBody))
	}

	parseOpenAISSEStream(resp.Body, ch)
	return nil
}

func (c *OpenAIClient) buildRequestBody(opts StreamOptions) ([]byte, error) {
	messages := convertToOpenAIMessages(opts.Messages, opts.System)

	req := openaiRequest{
		Model:       opts.Model,
		Messages:    messages,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
		Stream:      true,
	}

	if len(opts.Tools) > 0 {
		req.Tools = convertToOpenAITools(opts.Tools)
	}

	return json.Marshal(req)
}

// newOpenAIFactory returns a ClientFactory for OpenAI-compatible providers.
func newOpenAIFactory(profile config.ProviderProfile, apiKey string) (Client, error) {
	opts := []OpenAIClientOption{WithOpenAIAPIKey(apiKey)}
	if profile.BaseURL != "" {
		opts = append(opts, WithOpenAIBaseURL(profile.BaseURL))
	}
	return NewOpenAIClient(opts...), nil
}
