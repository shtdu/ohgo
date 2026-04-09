package ui

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWriter struct {
	buf bytes.Buffer
}

func (m *mockWriter) Printf(format string, args ...any) {
	m.buf.WriteString(fmt.Sprintf(format, args...))
}

func TestPermissionPrompter_Allow(t *testing.T) {
	input := bufio.NewReader(strings.NewReader("y\n"))
	w := &mockWriter{}
	p := NewPermissionPrompter(input, w)

	allow, remember, err := p.PromptApproval(context.Background(), "bash", "ls -la")
	require.NoError(t, err)
	assert.True(t, allow)
	assert.False(t, remember)
}

func TestPermissionPrompter_Deny(t *testing.T) {
	input := bufio.NewReader(strings.NewReader("n\n"))
	w := &mockWriter{}
	p := NewPermissionPrompter(input, w)

	allow, remember, err := p.PromptApproval(context.Background(), "write_file", "")
	require.NoError(t, err)
	assert.False(t, allow)
	assert.False(t, remember)
}

func TestPermissionPrompter_Always(t *testing.T) {
	input := bufio.NewReader(strings.NewReader("always\n"))
	w := &mockWriter{}
	p := NewPermissionPrompter(input, w)

	allow, remember, err := p.PromptApproval(context.Background(), "bash", "rm -rf")
	require.NoError(t, err)
	assert.True(t, allow)
	assert.True(t, remember)
}

func TestPermissionPrompter_ContextCancel(t *testing.T) {
	input := bufio.NewReader(strings.NewReader("")) // no input
	w := &mockWriter{}
	p := NewPermissionPrompter(input, w)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := p.PromptApproval(ctx, "bash", "")
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestPermissionPrompter_UnknownAnswer(t *testing.T) {
	input := bufio.NewReader(strings.NewReader("maybe\n"))
	w := &mockWriter{}
	p := NewPermissionPrompter(input, w)

	allow, _, err := p.PromptApproval(context.Background(), "bash", "")
	require.NoError(t, err)
	assert.False(t, allow) // unknown answers are treated as deny
}
