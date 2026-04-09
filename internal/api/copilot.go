package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/shtdu/ohgo/internal/config"
)

const (
	copilotTokenURL = "https://api.github.com/copilot_internal/v2/token"
	copilotBaseURL  = "https://api.githubcopilot.com/chat/completions"
)

// CopilotClient implements Client for GitHub Copilot's API.
// It handles two-step auth: OAuth token → Copilot API token → API request.
type CopilotClient struct {
	oauthToken  string
	baseURL     string
	tokenURL    string
	httpClient  *http.Client
	maxRetries  int

	tokenMu     sync.RWMutex
	copilotToken string
	tokenExpiry  time.Time
}

// CopilotClientOption configures a CopilotClient.
type CopilotClientOption func(*CopilotClient)

// WithCopilotOAuthToken sets the GitHub OAuth token.
func WithCopilotOAuthToken(token string) CopilotClientOption {
	return func(c *CopilotClient) { c.oauthToken = token }
}

// WithCopilotBaseURL sets a custom API base URL.
func WithCopilotBaseURL(url string) CopilotClientOption {
	return func(c *CopilotClient) { c.baseURL = url }
}

// WithCopilotHTTPClient sets a custom HTTP client.
func WithCopilotHTTPClient(hc *http.Client) CopilotClientOption {
	return func(c *CopilotClient) { c.httpClient = hc }
}

// WithCopilotMaxRetries sets the maximum number of retries.
func WithCopilotMaxRetries(n int) CopilotClientOption {
	return func(c *CopilotClient) { c.maxRetries = n }
}

// NewCopilotClient creates a new GitHub Copilot API client.
func NewCopilotClient(opts ...CopilotClientOption) *CopilotClient {
	c := &CopilotClient{
		baseURL:    copilotBaseURL,
		tokenURL:   copilotTokenURL,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
		maxRetries: maxRetries,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Stream sends a request via the Copilot API and returns a channel of events.
func (c *CopilotClient) Stream(ctx context.Context, opts StreamOptions) (<-chan StreamEvent, error) {
	ch := make(chan StreamEvent, 64)

	go func() {
		defer close(ch)
		c.streamWithRetry(ctx, opts, ch)
	}()

	return ch, nil
}

func (c *CopilotClient) streamWithRetry(ctx context.Context, opts StreamOptions, ch chan<- StreamEvent) {
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := c.streamOnce(ctx, opts, ch); err != nil {
			if !IsRetryable(err) || attempt >= c.maxRetries {
				ch <- StreamEvent{Type: "error", Data: err.Error()}
				return
			}
			delay := retryDelay(attempt)
			slog.Warn("copilot retry", "attempt", attempt+1, "max", c.maxRetries, "delay", delay, "error", err)
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

func (c *CopilotClient) streamOnce(ctx context.Context, opts StreamOptions, ch chan<- StreamEvent) error {
	// Step 1: Get/refresh Copilot API token.
	apiToken, err := c.getCopilotToken(ctx)
	if err != nil {
		return fmt.Errorf("copilot token: %w", err)
	}

	// Step 2: Build request.
	body, err := c.buildRequestBody(opts)
	if err != nil {
		return fmt.Errorf("build request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Accept", "text/event-stream")
	// Copilot-specific headers.
	req.Header.Set("Editor-Version", "ohgo/1.0")
	req.Header.Set("Editor-Plugin-Version", "ohgo/1.0")
	req.Header.Set("Openai-Organization", "")

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

func (c *CopilotClient) getCopilotToken(ctx context.Context) (string, error) {
	c.tokenMu.RLock()
	if c.copilotToken != "" && time.Now().Before(c.tokenExpiry) {
		token := c.copilotToken
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()

	// Fetch new token.
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// Double-check after acquiring write lock.
	if c.copilotToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.copilotToken, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.tokenURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.oauthToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", &APIError{StatusCode: 0, Message: err.Error(), Retryable: true}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", &APIError{StatusCode: resp.StatusCode, Message: string(body), Retryable: false}
	}

	var tokenResp struct {
		Token     string `json:"token"`
		ExpiresAt int64  `json:"expires_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	c.copilotToken = tokenResp.Token
	c.tokenExpiry = time.Unix(tokenResp.ExpiresAt, 0).Add(-30 * time.Second) // 30s buffer

	return c.copilotToken, nil
}

func (c *CopilotClient) buildRequestBody(opts StreamOptions) ([]byte, error) {
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

// newCopilotFactory returns a ClientFactory for the Copilot provider.
func newCopilotFactory(profile config.ProviderProfile, apiKey string) (Client, error) {
	opts := []CopilotClientOption{WithCopilotOAuthToken(apiKey)}
	if profile.BaseURL != "" {
		opts = append(opts, WithCopilotBaseURL(profile.BaseURL))
	}
	return NewCopilotClient(opts...), nil
}
