package hooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefinitionExecutor_CommandHookSuccess(t *testing.T) {
	reg := NewRegistry()
	reg.Register(HookEventPreToolUse, HookDefinition{
		Event:   HookEventPreToolUse,
		Type:    HookTypeCommand,
		Command: "echo hello",
	})

	exec := NewDefinitionExecutor(reg)
	agg, err := exec.ExecutePre(context.Background(), "read_file", nil)
	require.NoError(t, err)
	require.Len(t, agg.Results, 1)
	assert.True(t, agg.Results[0].Success)
	assert.Contains(t, agg.Results[0].Output, "hello")
	assert.False(t, agg.Results[0].Blocked)
}

func TestDefinitionExecutor_CommandHookFailureWithBlock(t *testing.T) {
	reg := NewRegistry()
	reg.Register(HookEventPreToolUse, HookDefinition{
		Event:          HookEventPreToolUse,
		Type:           HookTypeCommand,
		Command:        "exit 1",
		BlockOnFailure: true,
	})

	exec := NewDefinitionExecutor(reg)
	agg, err := exec.ExecutePre(context.Background(), "bash", nil)
	require.NoError(t, err)
	require.Len(t, agg.Results, 1)
	assert.False(t, agg.Results[0].Success)
	assert.True(t, agg.Results[0].Blocked)
	assert.Contains(t, agg.Results[0].Reason, "command hook failed")
}

func TestDefinitionExecutor_CommandHookFailureWithoutBlock(t *testing.T) {
	reg := NewRegistry()
	reg.Register(HookEventPreToolUse, HookDefinition{
		Event:          HookEventPreToolUse,
		Type:           HookTypeCommand,
		Command:        "exit 1",
		BlockOnFailure: false,
	})

	exec := NewDefinitionExecutor(reg)
	agg, err := exec.ExecutePre(context.Background(), "bash", nil)
	require.NoError(t, err)
	require.Len(t, agg.Results, 1)
	assert.False(t, agg.Results[0].Success)
	assert.False(t, agg.Results[0].Blocked)
}

func TestDefinitionExecutor_HTTPHookSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	reg := NewRegistry()
	reg.Register(HookEventPostToolUse, HookDefinition{
		Event: HookEventPostToolUse,
		Type:  HookTypeHTTP,
		URL:   server.URL,
	})

	exec := NewDefinitionExecutor(reg)
	agg, err := exec.ExecutePost(context.Background(), "read_file", nil, "file contents")
	require.NoError(t, err)
	require.Len(t, agg.Results, 1)
	assert.True(t, agg.Results[0].Success)
	assert.False(t, agg.Results[0].Blocked)
}

func TestDefinitionExecutor_HTTPHookFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	reg := NewRegistry()
	reg.Register(HookEventPreToolUse, HookDefinition{
		Event:          HookEventPreToolUse,
		Type:           HookTypeHTTP,
		URL:            server.URL,
		BlockOnFailure: true,
	})

	exec := NewDefinitionExecutor(reg)
	agg, err := exec.ExecutePre(context.Background(), "bash", nil)
	require.NoError(t, err)
	require.Len(t, agg.Results, 1)
	assert.False(t, agg.Results[0].Success)
	assert.True(t, agg.Results[0].Blocked)
	assert.Contains(t, agg.Results[0].Reason, "HTTP 500")
}

func TestDefinitionExecutor_PreHookBlockStopsSubsequent(t *testing.T) {
	var secondCalled int32

	reg := NewRegistry()
	// First hook blocks
	reg.Register(HookEventPreToolUse, HookDefinition{
		Event:          HookEventPreToolUse,
		Type:           HookTypeCommand,
		Command:        "exit 1",
		BlockOnFailure: true,
	})
	// Second hook should never run
	reg.Register(HookEventPreToolUse, HookDefinition{
		Event:   HookEventPreToolUse,
		Type:    HookTypeCommand,
		Command: "sh -c 'echo second_ran'",
	})

	exec := NewDefinitionExecutor(reg)
	agg, err := exec.ExecutePre(context.Background(), "bash", nil)
	require.NoError(t, err)

	// Only one result because second hook should not have run
	assert.Len(t, agg.Results, 1)
	assert.True(t, agg.Results[0].Blocked)
	assert.Equal(t, int32(0), atomic.LoadInt32(&secondCalled))
}

func TestDefinitionExecutor_MatcherFiltering(t *testing.T) {
	reg := NewRegistry()
	// Hook only matches "bash"
	reg.Register(HookEventPreToolUse, HookDefinition{
		Event:   HookEventPreToolUse,
		Type:    HookTypeCommand,
		Matcher: "bash",
		Command: "echo matched",
	})

	exec := NewDefinitionExecutor(reg)

	// Should not match "read_file"
	agg, err := exec.ExecutePre(context.Background(), "read_file", nil)
	require.NoError(t, err)
	assert.Empty(t, agg.Results)

	// Should match "bash"
	agg, err = exec.ExecutePre(context.Background(), "bash", nil)
	require.NoError(t, err)
	assert.Len(t, agg.Results, 1)
	assert.True(t, agg.Results[0].Success)
}

func TestDefinitionExecutor_ContextCancellation(t *testing.T) {
	reg := NewRegistry()
	// Hook that sleeps long enough to be cancelled
	reg.Register(HookEventPreToolUse, HookDefinition{
		Event:   HookEventPreToolUse,
		Type:    HookTypeCommand,
		Command: "sleep 10",
	})

	exec := NewDefinitionExecutor(reg)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	agg, err := exec.ExecutePre(ctx, "bash", nil)
	require.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	// May or may not have results depending on timing
	if len(agg.Results) > 0 {
		assert.False(t, agg.Results[0].Success)
	}
}

func TestDefinitionExecutor_UpdateRegistry(t *testing.T) {
	// Start with one registry containing no hooks
	reg1 := NewRegistry()
	exec := NewDefinitionExecutor(reg1)

	agg, err := exec.ExecutePre(context.Background(), "bash", nil)
	require.NoError(t, err)
	assert.Empty(t, agg.Results)

	// Update to a new registry with a hook
	reg2 := NewRegistry()
	reg2.Register(HookEventPreToolUse, HookDefinition{
		Event:   HookEventPreToolUse,
		Type:    HookTypeCommand,
		Command: "echo updated",
	})
	exec.UpdateRegistry(reg2)

	agg, err = exec.ExecutePre(context.Background(), "bash", nil)
	require.NoError(t, err)
	require.Len(t, agg.Results, 1)
	assert.True(t, agg.Results[0].Success)
	assert.Contains(t, agg.Results[0].Output, "updated")
}
