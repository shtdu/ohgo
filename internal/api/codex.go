package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/shtdu/ohgo/internal/config"
)

const defaultCodexBaseURL = "http://localhost:8967/v1/chat/completions"

// CodexClient implements Client for OpenAI Codex CLI's local API.
type CodexClient struct {
	token     string
	baseURL   string
	httpClient *http.Client
	maxRetries int
}

// CodexClientOption configures a CodexClient.
type CodexClientOption func(*CodexClient)

// WithCodexToken sets the Codex API token.
func WithCodexToken(token string) CodexClientOption {
	return func(c *CodexClient) { c.token = token }
}

// WithCodexBaseURL sets a custom API base URL.
func WithCodexBaseURL(url string) CodexClientOption {
	return func(c *CodexClient) { c.baseURL = url }
}

// WithCodexHTTPClient sets a custom HTTP client.
func WithCodexHTTPClient(hc *http.Client) CodexClientOption {
	return func(c *CodexClient) { c.httpClient = hc }
}

// WithCodexMaxRetries sets the maximum number of retries.
func WithCodexMaxRetries(n int) CodexClientOption {
	return func(c *CodexClient) { c.maxRetries = n }
}

// NewCodexClient creates a new Codex API client.
func NewCodexClient(opts ...CodexClientOption) *CodexClient {
	c := &CodexClient{
		baseURL:    defaultCodexBaseURL,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
		maxRetries: maxRetries,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Stream sends a request to the Codex API and returns a channel of events.
func (c *CodexClient) Stream(ctx context.Context, opts StreamOptions) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 64)

	go func() {
		defer close(ch)
		c.streamWithRetry(ctx, opts, ch)
	}()

	return ch, nil
}

func (c *CodexClient) streamWithRetry(ctx context.Context, opts StreamOptions, ch chan<- StreamEvent) {
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := c.streamOnce(ctx, opts, ch); err != nil {
			if !IsRetryable(err) || attempt >= c.maxRetries {
				ch <- StreamEvent{Type: "error", Data: err.Error()}
				return
			}
			delay := retryDelay(attempt)
			log.Printf("codex retry attempt %d/%d after %s: %v", attempt+1, c.maxRetries, delay, err)
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

func (c *CodexClient) streamOnce(ctx context.Context, opts StreamOptions, ch chan<- StreamEvent) error {
	body, err := c.buildRequestBody(opts)
	if err != nil {
		return fmt.Errorf("build request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
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

func (c *CodexClient) buildRequestBody(opts StreamOptions) ([]byte, error) {
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

// extractCodexToken reads the Codex CLI's local credentials.
// Returns (token, baseURL, error).
func extractCodexToken() (string, string, error) {
	// Check environment variables first.
	if token := os.Getenv("CODEX_TOKEN"); token != "" {
		baseURL := os.Getenv("CODEX_API_URL")
		if baseURL == "" {
			baseURL = defaultCodexBaseURL
		}
		return token, baseURL, nil
	}

	// Fall back to ~/.codex/credentials.json.
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("codex: find home dir: %w", err)
	}

	credPath := filepath.Join(home, ".codex", "credentials.json")
	data, err := os.ReadFile(credPath)
	if err != nil {
		return "", "", fmt.Errorf("codex: read credentials: %w (install Codex CLI or set CODEX_TOKEN)", err)
	}

	var credData map[string]any
	if err := json.Unmarshal(data, &credData); err != nil {
		return "", "", fmt.Errorf("codex: parse credentials: %w", err)
	}

	token, _ := credData["token"].(string)
	if token == "" {
		token, _ = credData["api_key"].(string)
	}
	if token == "" {
		return "", "", fmt.Errorf("codex: no token found in %s", credPath)
	}

	baseURL := defaultCodexBaseURL
	if url, ok := credData["api_url"].(string); ok && url != "" {
		baseURL = url
	}

	return token, baseURL, nil
}

// newCodexFactory returns a ClientFactory for the Codex provider.
func newCodexFactory(profile config.ProviderProfile, apiKey string) (Client, error) {
	token := apiKey
	baseURL := profile.BaseURL

	// If no explicit key, try extracting from Codex CLI config.
	if token == "" {
		extractedToken, extractedURL, err := extractCodexToken()
		if err != nil {
			return nil, fmt.Errorf("codex: %w", err)
		}
		token = extractedToken
		if baseURL == "" {
			baseURL = extractedURL
		}
	}

	opts := []CodexClientOption{WithCodexToken(token)}
	if baseURL != "" {
		opts = append(opts, WithCodexBaseURL(baseURL))
	}
	return NewCodexClient(opts...), nil
}
