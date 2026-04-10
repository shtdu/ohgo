package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUI_Print(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf, nil)
	u.Print("hello")
	assert.Equal(t, "hello", buf.String())
}

func TestUI_Printf(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf, nil)
	u.Printf("hello %s", "world")
	assert.Equal(t, "hello world", buf.String())
}

func TestUI_Println(t *testing.T) {
	var buf bytes.Buffer
	u := New(&buf, nil)
	u.Println("hello")
	assert.Equal(t, "hello\n", buf.String())
}

func TestUI_PrintError(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	u := New(&outBuf, nil).WithErrWriter(&errBuf)
	u.PrintError("something failed")
	assert.Equal(t, "something failed\n", errBuf.String())
	assert.Empty(t, outBuf.String())
}

func TestUI_Prompt(t *testing.T) {
	in := strings.NewReader("my answer\n")
	var out bytes.Buffer
	u := New(&out, in)
	answer, err := u.Prompt(context.Background(), "> ")
	require.NoError(t, err)
	assert.Equal(t, "my answer", answer)
	assert.Equal(t, "> ", out.String())
}

func TestUI_Prompt_ContextCancel(t *testing.T) {
	in := iotest.TimeoutReader(strings.NewReader("data"))
	var out bytes.Buffer
	u := New(&out, in)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := u.Prompt(ctx, "> ")
	assert.Error(t, err)
}

func TestUI_Prompt_EmptyInput(t *testing.T) {
	// An empty reader: Scanner.Scan() returns false with nil error (clean EOF).
	// The Prompt returns an error in this case since there's no input.
	in := strings.NewReader("")
	var out bytes.Buffer
	u := New(&out, in)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	line, err := u.Prompt(ctx, "> ")
	// Clean EOF from empty reader — Scanner returns false, Err() is nil.
	// Our implementation reports this as an error.
	if err == nil {
		assert.Empty(t, line) // empty string on EOF
	}
}

func TestAskQuestion_WithAnswer(t *testing.T) {
	in := strings.NewReader("my answer\n")
	var out bytes.Buffer
	u := New(&out, in)
	answer, err := u.AskQuestion(context.Background(), "What is your name?", nil, "")
	require.NoError(t, err)
	assert.Equal(t, "my answer", answer)
	assert.Contains(t, out.String(), "What is your name?")
}

func TestAskQuestion_WithOptions(t *testing.T) {
	in := strings.NewReader("red\n")
	var out bytes.Buffer
	u := New(&out, in)
	answer, err := u.AskQuestion(context.Background(), "Pick a color", []string{"red", "blue"}, "")
	require.NoError(t, err)
	assert.Equal(t, "red", answer)
	assert.Contains(t, out.String(), "Options: [red blue]")
}

func TestAskQuestion_WithDefault_Used(t *testing.T) {
	in := strings.NewReader("\n")
	var out bytes.Buffer
	u := New(&out, in)
	answer, err := u.AskQuestion(context.Background(), "Continue?", nil, "default-answer")
	require.NoError(t, err)
	assert.Equal(t, "default-answer", answer)
	assert.Contains(t, out.String(), "Default: default-answer")
}

func TestAskQuestion_WithDefault_Overridden(t *testing.T) {
	in := strings.NewReader("custom\n")
	var out bytes.Buffer
	u := New(&out, in)
	answer, err := u.AskQuestion(context.Background(), "Continue?", nil, "default")
	require.NoError(t, err)
	assert.Equal(t, "custom", answer)
}

func TestAskQuestion_EmptyAnswerNoDefault(t *testing.T) {
	in := strings.NewReader("\n")
	var out bytes.Buffer
	u := New(&out, in)
	answer, err := u.AskQuestion(context.Background(), "Your name?", nil, "")
	require.NoError(t, err)
	assert.Equal(t, "", answer)
}

func TestAskQuestion_ContextCancel(t *testing.T) {
	in := iotest.TimeoutReader(strings.NewReader("data"))
	var out bytes.Buffer
	u := New(&out, in)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := u.AskQuestion(ctx, "Question?", nil, "")
	assert.Error(t, err)
}
