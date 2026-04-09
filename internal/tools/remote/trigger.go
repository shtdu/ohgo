// Package remote provides tools for triggering remote actions via HTTP.
package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shtdu/ohgo/internal/tools"
)

const maxResponseBody = 5000

type triggerInput struct {
	URL             string            `json:"url"`
	Method          string            `json:"method,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	Body            string            `json:"body,omitempty"`
	TimeoutSeconds  int               `json:"timeout_seconds,omitempty"`
}

// RemoteTriggerTool sends an HTTP request to a specified URL and returns
// the response status and body.
type RemoteTriggerTool struct {
	// Client is the HTTP client used for requests. Defaults to http.DefaultClient if nil.
	Client *http.Client
}

func (t RemoteTriggerTool) client() *http.Client {
	if t.Client != nil {
		return t.Client
	}
	return http.DefaultClient
}

func (RemoteTriggerTool) Name() string { return "remote_trigger" }

func (RemoteTriggerTool) Description() string {
	return "Triggers a remote action by sending an HTTP request to the specified URL. Returns the response status and body."
}

func (RemoteTriggerTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "The URL to send the request to (must be http or https)",
			},
			"method": map[string]any{
				"type":        "string",
				"description": "HTTP method (default: POST)",
				"default":     "POST",
			},
			"headers": map[string]any{
				"type":        "object",
				"description": "Optional HTTP headers to include in the request",
				"additionalProperties": map[string]any{"type": "string"},
			},
			"body": map[string]any{
				"type":        "string",
				"description": "Optional request body",
			},
			"timeout_seconds": map[string]any{
				"type":        "integer",
				"description": "Request timeout in seconds (default: 30)",
				"default":     30,
			},
		},
		"required":             []string{"url"},
		"additionalProperties": false,
	}
}

func (t RemoteTriggerTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	var input triggerInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.URL == "" {
		return tools.Result{Content: "url is required", IsError: true}, nil
	}

	parsedURL, err := url.Parse(input.URL)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid url: %v", err), IsError: true}, nil
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return tools.Result{
			Content: fmt.Sprintf("invalid url scheme %q: only http and https are supported", parsedURL.Scheme),
			IsError: true,
		}, nil
	}

	method := input.Method
	if method == "" {
		method = http.MethodPost
	}
	method = strings.ToUpper(method)

	timeout := time.Duration(input.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var bodyReader io.Reader
	if input.Body != "" {
		bodyReader = strings.NewReader(input.Body)
	}

	req, err := http.NewRequestWithContext(ctx, method, input.URL, bodyReader)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to create request: %v", err), IsError: true}, nil
	}

	for k, v := range input.Headers {
		req.Header.Set(k, v)
	}

	resp, err := t.client().Do(req)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("request failed: %v", err), IsError: true}, nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("failed to read response: %v", err), IsError: true}, nil
	}

	bodyStr := string(respBody)
	if len(bodyStr) > maxResponseBody {
		bodyStr = bodyStr[:maxResponseBody] + fmt.Sprintf("\n... (truncated, %d bytes total)", len(respBody))
	}

	return tools.Result{
		Content: fmt.Sprintf("Status: %d %s\n\n%s", resp.StatusCode, resp.Status, bodyStr),
	}, nil
}
