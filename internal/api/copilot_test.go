package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/config"
)

func TestCopilotClient_TextStreaming(t *testing.T) {
	var tokenRequests int32
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenRequests, 1)
		assert.Equal(t, "Bearer gh-oauth-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"token":"tid=copilot-token;exp=%d","expires_at":%d}`, time.Now().Add(10*time.Minute).Unix(), time.Now().Add(10*time.Minute).Unix()) //nolint:errcheck
	}))
	defer tokenServer.Close()

	sseData := `data: {"id":"chatcmpl-1","choices":[{"index":0,"delta":{"content":"Hello from Copilot!"},"finish_reason":null}]}

data: [DONE]
`
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer tid=copilot-token;exp="+fmt.Sprintf("%d", time.Now().Add(10*time.Minute).Unix()), r.Header.Get("Authorization"))
		assert.Equal(t, "ohgo/1.0", r.Header.Get("Editor-Version"))
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, sseData)
	}))
	defer apiServer.Close()

	client := NewCopilotClient(
		WithCopilotOAuthToken("gh-oauth-token"),
		WithCopilotBaseURL(apiServer.URL),
	)
	client.tokenURL = tokenServer.URL

	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	var textParts []string
	for e := range ch {
		if e.Type == "text_delta" {
			textParts = append(textParts, e.Data.(string))
		}
	}
	assert.Equal(t, []string{"Hello from Copilot!"}, textParts)
}

func TestCopilotClient_TokenCaching(t *testing.T) {
	var tokenRequests int32
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenRequests, 1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"token":"cached-token","expires_at":%d}`, time.Now().Add(10*time.Minute).Unix()) //nolint:errcheck
	}))
	defer tokenServer.Close()

	var apiRequests int32
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&apiRequests, 1)
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, `data: {"choices":[{"delta":{"content":"ok"}}]}\n\ndata: [DONE]\n`)
	}))
	defer apiServer.Close()

	client := NewCopilotClient(
		WithCopilotOAuthToken("gh-token"),
		WithCopilotBaseURL(apiServer.URL),
	)
	client.tokenURL = tokenServer.URL

	// Two streaming requests should only fetch the token once (cached).
	for i := 0; i < 2; i++ {
		ch, err := client.Stream(context.Background(), StreamOptions{
			Model:     "gpt-4",
			MaxTokens: 100,
			Messages:  []Message{NewUserTextMessage("hi")},
		})
		require.NoError(t, err)
		for range ch {}
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&tokenRequests), "token should be fetched only once")
	assert.Equal(t, int32(2), atomic.LoadInt32(&apiRequests))
}

func TestCopilotClient_AuthError(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, "unauthorized")
	}))
	defer tokenServer.Close()

	client := NewCopilotClient(
		WithCopilotOAuthToken("bad-token"),
	)
	client.tokenURL = tokenServer.URL

	ch, err := client.Stream(context.Background(), StreamOptions{
		Model:     "gpt-4",
		MaxTokens: 100,
		Messages:  []Message{NewUserTextMessage("hi")},
	})
	require.NoError(t, err)

	var events []StreamEvent
	for e := range ch {
		events = append(events, e)
	}
	require.Len(t, events, 1)
	assert.Equal(t, "error", events[0].Type)
}

func TestCopilotClient_Interface(t *testing.T) {
	var _ Client = (*CopilotClient)(nil)
}

func TestRegistry_CopilotFactory(t *testing.T) {
	r := NewRegistry()
	cfg := &config.Settings{
		APIKey: "gh-oauth-token",
		Profiles: map[string]config.ProviderProfile{
			"copilot": {
				APIFormat: "copilot",
			},
		},
		ActiveProfile: "copilot",
	}

	client, err := r.CreateClient(cfg, "")
	require.NoError(t, err)
	cc, ok := client.(*CopilotClient)
	require.True(t, ok)
	assert.Equal(t, "gh-oauth-token", cc.oauthToken)
}
