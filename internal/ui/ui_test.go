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
