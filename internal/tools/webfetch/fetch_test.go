package webfetch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestFetchTool_Name(t *testing.T) {
	assert.Equal(t, "web_fetch", FetchTool{}.Name())
}

func TestFetchTool_HTMLContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><head><title>Test</title><style>body{}</style></head><body><h1>Hello</h1><p>World</p></body></html>`))
	}))
	defer server.Close()

	tool := FetchTool{}
	args, _ := json.Marshal(map[string]any{"url": server.URL, "max_chars": 5000})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Hello")
	assert.Contains(t, result.Content, "World")
	assert.NotContains(t, result.Content, "<html>")
	assert.NotContains(t, result.Content, "<style>")
}

func TestFetchTool_JSONContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key": "value"}`))
	}))
	defer server.Close()

	tool := FetchTool{}
	args, _ := json.Marshal(map[string]string{"url": server.URL})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, `"key": "value"`)
}

func TestFetchTool_404Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tool := FetchTool{}
	args, _ := json.Marshal(map[string]string{"url": server.URL})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "404")
}

func TestFetchTool_InvalidURL(t *testing.T) {
	tool := FetchTool{}
	args, _ := json.Marshal(map[string]string{"url": "not-a-url"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "http://")
}

func TestFetchTool_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	tool := FetchTool{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	args, _ := json.Marshal(map[string]string{"url": server.URL})
	result, err := tool.Execute(ctx, args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "fetch error")
}

func TestFetchTool_Truncation(t *testing.T) {
	longBody := ""
	for i := 0; i < 500; i++ {
		longBody += "abcdefghij"
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(longBody))
	}))
	defer server.Close()

	tool := FetchTool{}
	args, _ := json.Marshal(map[string]any{"url": server.URL, "max_chars": 500})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "[truncated]")
}

func TestFetchTool_InvalidJSON(t *testing.T) {
	tool := FetchTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

var _ tools.Tool = FetchTool{}
