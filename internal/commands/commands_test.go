package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubCmd is a minimal Command for testing.
type stubCmd struct {
	name      string
	shortHelp string
}

func (s stubCmd) Name() string        { return s.name }
func (s stubCmd) ShortHelp() string   { return s.shortHelp }
func (s stubCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: "ok"}, nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	c := stubCmd{name: "test", shortHelp: "test command"}
	r.Register(c)
	got := r.Get("test")
	require.NotNil(t, got)
	assert.Equal(t, "test", got.Name())
}

func TestRegistry_GetMissing(t *testing.T) {
	r := NewRegistry()
	assert.Nil(t, r.Get("nonexistent"))
}

func TestRegistry_Lookup(t *testing.T) {
	r := NewRegistry()
	r.Register(stubCmd{name: "help", shortHelp: "show help"})
	r.Register(stubCmd{name: "exit", shortHelp: "exit repl"})

	tests := []struct {
		line      string
		wantName  string
		wantArgs  string
		wantFound bool
	}{
		{"/help", "help", "", true},
		{"/help arg1 arg2", "help", "arg1 arg2", true},
		{"/exit", "exit", "", true},
		{"/unknown", "", "", false},
		{"not a slash command", "", "", false},
		{"", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			cmd, args, ok := r.Lookup(tt.line)
			if !tt.wantFound {
				assert.False(t, ok)
				return
			}
			require.True(t, ok)
			assert.Equal(t, tt.wantName, cmd.Name())
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(stubCmd{name: "zebra", shortHelp: "z"})
	r.Register(stubCmd{name: "alpha", shortHelp: "a"})
	r.Register(stubCmd{name: "middle", shortHelp: "m"})

	list := r.List()
	require.Len(t, list, 3)
	// Sorted by name.
	assert.Equal(t, "alpha", list[0].Name())
	assert.Equal(t, "middle", list[1].Name())
	assert.Equal(t, "zebra", list[2].Name())
}
