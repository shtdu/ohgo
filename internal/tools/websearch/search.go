// Package websearch implements the web_search tool using DuckDuckGo HTML search.
package websearch

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
	"github.com/shtdu/ohgo/internal/tools/htmlutil"
)

const (
	defaultMaxResults = 5
	maxMaxResults     = 10
	searchTimeout     = 20 * time.Second
	ddgSearchURL      = "https://html.duckduckgo.com/html/"
	searchUserAgent   = "og/0.1"
)

type searchInput struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results"`
}

type searchResult struct {
	Title   string
	URL     string
	Snippet string
}

// SearchTool searches the web using DuckDuckGo.
type SearchTool struct {
	Client *http.Client
	// SearchURL overrides the default DuckDuckGo endpoint (for testing).
	SearchURL string
}

func (SearchTool) Name() string { return "web_search" }

func (SearchTool) Description() string {
	return "Search the web using DuckDuckGo. Returns results with titles, URLs, and snippets."
}

func (SearchTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Search query",
			},
			"max_results": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results",
				"default":     defaultMaxResults,
				"minimum":     1,
				"maximum":     maxMaxResults,
			},
		},
		"required":             []string{"query"},
		"additionalProperties": false,
	}
}

func (s SearchTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input searchInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	if input.Query == "" {
		return tools.Result{Content: "query is required", IsError: true}, nil
	}

	maxResults := input.MaxResults
	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}
	maxResults = min(maxResults, maxMaxResults)

	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: searchTimeout}
	}

	searchURL := s.SearchURL
	if searchURL == "" {
		searchURL = ddgSearchURL
	}

	ctx, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()

	// POST to DuckDuckGo HTML search
	formData := url.Values{}
	formData.Set("q", input.Query)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, searchURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("search error: %v", err), IsError: true}, nil
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", searchUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("search error: %v", err), IsError: true}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return tools.Result{Content: fmt.Sprintf("search HTTP %d", resp.StatusCode), IsError: true}, nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("read error: %v", err), IsError: true}, nil
	}

	results := parseDDGResults(string(body), maxResults)
	if len(results) == 0 {
		return tools.Result{Content: "No search results found"}, nil
	}

	var buf strings.Builder
	for i, r := range results {
		fmt.Fprintf(&buf, "%d. %s\n   %s\n   %s\n", i+1, r.Title, r.URL, r.Snippet)
		if i < len(results)-1 {
			buf.WriteString("\n")
		}
	}

	return tools.Result{Content: buf.String()}, nil
}

// parseDDGResults extracts search results from DuckDuckGo HTML.
func parseDDGResults(html string, maxResults int) []searchResult {
	var results []searchResult

	// Find result blocks by locating <a class="result__a" elements
	// DDG uses this class for result links
	remaining := html
	for len(results) < maxResults {
		idx := strings.Index(remaining, `class="result__a"`)
		if idx < 0 {
			break
		}

		// Extract the block from this result to the next one (or end)
		block := remaining[idx:]
		nextResult := strings.Index(block[len(`class="result__a"`):], `class="result__a"`)
		if nextResult > 0 {
			block = block[:len(`class="result__a"`)+nextResult]
			remaining = remaining[idx+len(`class="result__a"`)+nextResult:]
		} else {
			remaining = ""
		}

		var r searchResult

		// Extract URL from href — look backwards from the class attribute
		preBlock := html[:strings.Index(html, `class="result__a"`)] // approximate
		_ = preBlock
		// Find href before or at the class attribute in the full tag
		href := extractAttr(block, "href")
		if href != "" {
			// DDG uses redirect URLs — extract the real URL from uddg param
			if u, err := url.Parse(href); err == nil {
				if uddg := u.Query().Get("uddg"); uddg != "" {
					r.URL = uddg
				} else {
					r.URL = href
				}
			}
		}

		// Extract title: text between > and </a> after class="result__a"
		r.Title = extractTagContent(block, `class="result__a"`)
		r.Title = strings.TrimSpace(decodeHTMLEntities(r.Title))

		// Extract snippet
		r.Snippet = extractTagContent(block, `class="result__snippet"`)
		r.Snippet = stripTags(r.Snippet)
		r.Snippet = strings.TrimSpace(decodeHTMLEntities(r.Snippet))

		if r.Title != "" || r.URL != "" {
			results = append(results, r)
		}
	}

	return results
}

// extractTagContent returns the text content after a tag attribute up to </a>.
func extractTagContent(s, attr string) string {
	i := strings.Index(s, attr)
	if i < 0 {
		return ""
	}
	s = s[i+len(attr):]
	// Find closing > of opening tag
	gt := strings.Index(s, ">")
	if gt < 0 {
		return ""
	}
	s = s[gt+1:]
	end := strings.Index(s, "</a>")
	if end < 0 {
		return s
	}
	return s[:end]
}

// extractAttr extracts an attribute value from a tag near the start of s.
func extractAttr(s, attr string) string {
	// Find href=" in the first ~500 chars (within the opening tag)
	search := s
	if len(search) > 500 {
		search = search[:500]
	}
	prefix := attr + `="`
	i := strings.Index(search, prefix)
	if i < 0 {
		return ""
	}
	val := search[i+len(prefix):]
	end := strings.Index(val, `"`)
	if end < 0 {
		return ""
	}
	return val[:end]
}

func stripTags(s string) string {
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
	return result.String()
}

func decodeHTMLEntities(s string) string {
	return htmlutil.DecodeEntities(s)
}
