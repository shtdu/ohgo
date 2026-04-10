package websearch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

const testDDGHTML = `
<html>
<body>
<div class="result">
<a rel="nofollow" class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com%2Fpage&rut=abc">Example Page Title</a>
<a class="result__snippet">This is a snippet about the example page.</a>
</div>
<div class="result">
<a rel="nofollow" class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fgolang.org&rut=def">The Go Programming Language</a>
<a class="result__snippet">Go is an open source programming language.</a>
</div>
<div class="result">
<a rel="nofollow" class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fgithub.com&rut=ghi">GitHub: Let&apos;s build from here</a>
<a class="result__snippet">The complete developer platform.</a>
</div>
</body>
</html>
`

func TestSearchTool_Name(t *testing.T) {
	assert.Equal(t, "web_search", SearchTool{}.Name())
}

func TestSearchTool_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(testDDGHTML))
	}))
	defer server.Close()

	tool := SearchTool{SearchURL: server.URL}
	args, _ := json.Marshal(map[string]any{"query": "golang", "max_results": 3})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Example Page Title")
	assert.Contains(t, result.Content, "https://example.com/page")
	assert.Contains(t, result.Content, "Go Programming Language")
}

func TestSearchTool_MaxResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testDDGHTML))
	}))
	defer server.Close()

	tool := SearchTool{SearchURL: server.URL}
	args, _ := json.Marshal(map[string]any{"query": "test", "max_results": 1})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "1.")
	assert.NotContains(t, result.Content, "2.")
}

func TestSearchTool_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body>No results here.</body></html>"))
	}))
	defer server.Close()

	tool := SearchTool{SearchURL: server.URL}
	args, _ := json.Marshal(map[string]string{"query": "obscurequery12345"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "No search results found")
}

func TestSearchTool_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	tool := SearchTool{SearchURL: server.URL}
	args, _ := json.Marshal(map[string]string{"query": "test"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "500")
}

func TestSearchTool_InvalidJSON(t *testing.T) {
	tool := SearchTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestSearchTool_EmptyQuery(t *testing.T) {
	tool := SearchTool{}
	args, _ := json.Marshal(map[string]string{"query": ""})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestParseDDGResults(t *testing.T) {
	results := parseDDGResults(testDDGHTML, 10)
	require.Len(t, results, 3)
	assert.Equal(t, "Example Page Title", results[0].Title)
	assert.Equal(t, "https://example.com/page", results[0].URL)
	assert.Contains(t, results[0].Snippet, "snippet about")
	assert.Contains(t, results[2].Title, "GitHub")
}

func TestSearchTool_HTMLEntityDecoding(t *testing.T) {
	results := parseDDGResults(testDDGHTML, 10)
	require.Len(t, results, 3)
	assert.True(t, strings.Contains(results[2].Title, "Let's"), "expected decoded entity, got: %s", results[2].Title)
}

var _ tools.Tool = SearchTool{}
