package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/tools"
)

// mustJSON marshals v to json.RawMessage.
func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

// ---- Interface satisfaction ----

var _ tools.Tool = RemoteTriggerTool{}

// ---- Tests ----

func TestRemoteTrigger_PostWithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "got: %s", string(body))
	}))
	defer srv.Close()

	tool := RemoteTriggerTool{}
	result, err := tool.Execute(context.Background(), mustJSON(t, map[string]any{
		"url":  srv.URL + "/hook",
		"body": `{"event":"deploy"}`,
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Status: 200")
	assert.Contains(t, result.Content, `got: {"event":"deploy"}`)
}

func TestRemoteTrigger_GetWithoutBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "hello")
	}))
	defer srv.Close()

	tool := RemoteTriggerTool{}
	result, err := tool.Execute(context.Background(), mustJSON(t, map[string]any{
		"url":    srv.URL + "/api",
		"method": "GET",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Status: 200")
	assert.Contains(t, result.Content, "hello")
}

func TestRemoteTrigger_CustomHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusAccepted)
		_, _ = fmt.Fprint(w, "accepted")
	}))
	defer srv.Close()

	tool := RemoteTriggerTool{}
	result, err := tool.Execute(context.Background(), mustJSON(t, map[string]any{
		"url":    srv.URL + "/webhook",
		"method": "POST",
		"headers": map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer test-token",
		},
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Status: 202")
	assert.Contains(t, result.Content, "accepted")
}

func TestRemoteTrigger_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tool := RemoteTriggerTool{}
	result, err := tool.Execute(context.Background(), mustJSON(t, map[string]any{
		"url":              srv.URL + "/slow",
		"timeout_seconds":  1,
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "request failed")
}

func TestRemoteTrigger_MissingURL(t *testing.T) {
	tool := RemoteTriggerTool{}
	result, err := tool.Execute(context.Background(), mustJSON(t, map[string]any{}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "url is required")
}

func TestRemoteTrigger_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tool := RemoteTriggerTool{}
	_, err := tool.Execute(ctx, mustJSON(t, map[string]any{
		"url": srv.URL + "/hook",
	}))
	assert.Error(t, err)
}

func TestRemoteTrigger_InvalidURLScheme(t *testing.T) {
	tool := RemoteTriggerTool{}

	result, err := tool.Execute(context.Background(), mustJSON(t, map[string]any{
		"url": "ftp://example.com/file",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid url scheme")
}

func TestRemoteTrigger_InvalidJSON(t *testing.T) {
	tool := RemoteTriggerTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestRemoteTrigger_ResponseTruncation(t *testing.T) {
	longBody := strings.Repeat("x", 10000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, longBody)
	}))
	defer srv.Close()

	tool := RemoteTriggerTool{}
	result, err := tool.Execute(context.Background(), mustJSON(t, map[string]any{
		"url":    srv.URL + "/big",
		"method": "GET",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "truncated")
	assert.Contains(t, result.Content, "10000 bytes total")
}

func TestRemoteTrigger_DefaultMethodIsPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "ok")
	}))
	defer srv.Close()

	tool := RemoteTriggerTool{}
	result, err := tool.Execute(context.Background(), mustJSON(t, map[string]any{
		"url": srv.URL + "/default",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Status: 200")
}
