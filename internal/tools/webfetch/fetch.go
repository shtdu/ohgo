// Package webfetch implements the web_fetch tool for fetching and reading web content.
package webfetch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shtdu/ohgo/internal/tools"
)

const (
	defaultMaxChars = 12000
	minMaxChars     = 500
	maxMaxChars     = 50000
	fetchTimeout    = 20 * time.Second
	userAgent       = "og/0.1"
)

type fetchInput struct {
	URL      string `json:"url"`
	MaxChars int    `json:"max_chars"`
}

// FetchTool fetches content from a URL and converts HTML to text.
type FetchTool struct {
	Client *http.Client
}

func (FetchTool) Name() string { return "web_fetch" }

func (FetchTool) Description() string {
	return "Fetches content from a URL. HTML pages are converted to plain text."
}

func (FetchTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url": map[string]any{
				"type":        "string",
				"description": "HTTP or HTTPS URL to fetch",
			},
			"max_chars": map[string]any{
				"type":        "integer",
				"description": "Maximum characters to return",
				"default":     defaultMaxChars,
				"minimum":     minMaxChars,
				"maximum":     maxMaxChars,
			},
		},
		"required":             []string{"url"},
		"additionalProperties": false,
	}
}

func (f FetchTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input fetchInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.URL == "" {
		return tools.Result{Content: "url is required", IsError: true}, nil
	}

	if !strings.HasPrefix(input.URL, "http://") && !strings.HasPrefix(input.URL, "https://") {
		return tools.Result{Content: "url must start with http:// or https://", IsError: true}, nil
	}

	maxChars := input.MaxChars
	if maxChars <= 0 {
		maxChars = defaultMaxChars
	}
	maxChars = clamp(maxChars, minMaxChars, maxMaxChars)

	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: fetchTimeout}
	}

	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, input.URL, nil)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid url: %v", err), IsError: true}, nil
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("fetch error: %v", err), IsError: true}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return tools.Result{Content: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status), IsError: true}, nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxChars+1000)))
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("read error: %v", err), IsError: true}, nil
	}

	content := string(body)
	contentType := resp.Header.Get("Content-Type")

	// Convert HTML to text if needed
	if strings.Contains(contentType, "text/html") {
		content = htmlToText(content)
	}

	// Truncate if needed
	truncated := false
	if len(content) > maxChars {
		content = content[:maxChars]
		truncated = true
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "URL: %s\nStatus: %s\nContent-Type: %s\n\n%s", input.URL, resp.Status, contentType, content)
	if truncated {
		buf.WriteString("\n...[truncated]")
	}

	return tools.Result{Content: buf.String()}, nil
}

// htmlToText performs basic HTML to plain text conversion.
func htmlToText(html string) string {
	s := html

	// Remove script and style blocks
	for _, tag := range []string{"script", "style"} {
		for {
			start := strings.Index(s, "<"+tag)
			if start == -1 {
				break
			}
			end := strings.Index(s, "</"+tag+">")
			if end == -1 || end <= start {
				break
			}
			s = s[:start] + s[end+len(tag)+3:]
		}
	}

	// Replace common block elements with newlines
	for _, tag := range []string{"br", "br/", "p", "div", "li", "tr", "h1", "h2", "h3", "h4", "h5", "h6"} {
		s = strings.ReplaceAll(s, "<"+tag+">", "\n")
		s = strings.ReplaceAll(s, "</"+tag+">", "\n")
	}

	// Remove all remaining HTML tags
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}

	// Decode common HTML entities
	text := result.String()
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")

	// Collapse whitespace
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return strings.Join(cleaned, "\n")
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}
