package hooks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookDefinitionUnmarshalCommand(t *testing.T) {
	raw := `{
		"event": "pre_tool_use",
		"type": "command",
		"matcher": "bash",
		"command": "check-script.sh",
		"timeout_seconds": 30
	}`
	var h HookDefinition
	err := json.Unmarshal([]byte(raw), &h)
	assert.NoError(t, err)
	assert.Equal(t, HookEventPreToolUse, h.Event)
	assert.Equal(t, HookTypeCommand, h.Type)
	assert.Equal(t, "bash", h.Matcher)
	assert.Equal(t, "check-script.sh", h.Command)
	assert.Equal(t, 30, h.TimeoutSeconds)
}

func TestHookDefinitionUnmarshalHTTP(t *testing.T) {
	raw := `{
		"event": "post_tool_use",
		"type": "http",
		"url": "https://example.com/hook",
		"headers": {"Authorization": "Bearer token123"}
	}`
	var h HookDefinition
	err := json.Unmarshal([]byte(raw), &h)
	assert.NoError(t, err)
	assert.Equal(t, HookEventPostToolUse, h.Event)
	assert.Equal(t, HookTypeHTTP, h.Type)
	assert.Equal(t, "https://example.com/hook", h.URL)
	assert.Equal(t, map[string]string{"Authorization": "Bearer token123"}, h.Headers)
}

func TestHookDefinitionUnmarshalPrompt(t *testing.T) {
	raw := `{
		"event": "pre_tool_use",
		"type": "prompt",
		"prompt": "Review this tool call for safety"
	}`
	var h HookDefinition
	err := json.Unmarshal([]byte(raw), &h)
	assert.NoError(t, err)
	assert.Equal(t, HookTypePrompt, h.Type)
	assert.Equal(t, "Review this tool call for safety", h.Prompt)
}

func TestHookDefinitionUnmarshalAgent(t *testing.T) {
	raw := `{
		"event": "post_tool_use",
		"type": "agent",
		"prompt": "Summarize the tool output",
		"model": "claude-sonnet-4-20250514"
	}`
	var h HookDefinition
	err := json.Unmarshal([]byte(raw), &h)
	assert.NoError(t, err)
	assert.Equal(t, HookTypeAgent, h.Type)
	assert.Equal(t, "Summarize the tool output", h.Prompt)
	assert.Equal(t, "claude-sonnet-4-20250514", h.Model)
}

func TestHookDefinitionValidate(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr string
	}{
		{
			name: "missing event",
			json: `{"type": "command", "command": "echo hi"}`,
			wantErr: "hook event is required",
		},
		{
			name: "missing type",
			json: `{"event": "pre_tool_use", "command": "echo hi"}`,
			wantErr: "hook type is required",
		},
		{
			name: "command hook without command",
			json: `{"event": "pre_tool_use", "type": "command"}`,
			wantErr: "command hook requires 'command' field",
		},
		{
			name: "http hook without url",
			json: `{"event": "pre_tool_use", "type": "http"}`,
			wantErr: "http hook requires 'url' field",
		},
		{
			name: "prompt hook without prompt",
			json: `{"event": "pre_tool_use", "type": "prompt"}`,
			wantErr: "prompt hook requires 'prompt' field",
		},
		{
			name: "agent hook without prompt",
			json: `{"event": "post_tool_use", "type": "agent"}`,
			wantErr: "agent hook requires 'prompt' field",
		},
		{
			name: "unknown hook type",
			json: `{"event": "pre_tool_use", "type": "unknown"}`,
			wantErr: "unknown hook type: unknown",
		},
		{
			name:    "valid command hook",
			json:    `{"event": "pre_tool_use", "type": "command", "command": "echo hi"}`,
			wantErr: "",
		},
		{
			name:    "valid http hook",
			json:    `{"event": "post_tool_use", "type": "http", "url": "https://example.com"}`,
			wantErr: "",
		},
		{
			name:    "valid prompt hook",
			json:    `{"event": "pre_tool_use", "type": "prompt", "prompt": "check safety"}`,
			wantErr: "",
		},
		{
			name:    "valid agent hook",
			json:    `{"event": "post_tool_use", "type": "agent", "prompt": "summarize"}`,
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h HookDefinition
			err := json.Unmarshal([]byte(tt.json), &h)
			assert.NoError(t, err)

			err = h.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}

func TestMatchesHook(t *testing.T) {
	tests := []struct {
		name    string
		matcher string
		subject string
		want    bool
	}{
		{
			name:    "exact match",
			matcher: "bash",
			subject: "bash",
			want:    true,
		},
		{
			name:    "glob pattern matches",
			matcher: "read_*",
			subject: "read_file",
			want:    true,
		},
		{
			name:    "glob pattern no match",
			matcher: "read_*",
			subject: "write_file",
			want:    false,
		},
		{
			name:    "empty matcher matches everything",
			matcher: "",
			subject: "anything",
			want:    true,
		},
		{
			name:    "single char glob",
			matcher: "bas?",
			subject: "bash",
			want:    true,
		},
		{
			name:    "non-matching exact",
			matcher: "bash",
			subject: "sh",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesHook(tt.matcher, tt.subject)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAggregatedResultBlocked(t *testing.T) {
	tests := []struct {
		name    string
		results []HookResult
		blocked bool
		reason  string
	}{
		{
			name:    "empty results",
			results: nil,
			blocked: false,
			reason:  "",
		},
		{
			name: "all passing",
			results: []HookResult{
				{Success: true, Blocked: false},
				{Success: true, Blocked: false},
			},
			blocked: false,
			reason:  "",
		},
		{
			name: "one blocked",
			results: []HookResult{
				{Success: true, Blocked: false},
				{Success: false, Blocked: true, Reason: "dangerous command"},
			},
			blocked: true,
			reason:  "dangerous command",
		},
		{
			name: "multiple blocked returns first reason",
			results: []HookResult{
				{Success: false, Blocked: true, Reason: "first block"},
				{Success: false, Blocked: true, Reason: "second block"},
			},
			blocked: true,
			reason:  "first block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := &AggregatedResult{Results: tt.results}
			assert.Equal(t, tt.blocked, ar.Blocked())
			assert.Equal(t, tt.reason, ar.Reason())
		})
	}
}
