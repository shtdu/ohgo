package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/shtdu/ohgo/internal/memory"
)

type memoryCmd struct{}

var _ Command = memoryCmd{}

func (memoryCmd) Name() string        { return "memory" }
func (memoryCmd) ShortHelp() string   { return "list memory entries" }

func (memoryCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	if deps.Cwd == "" {
		return Result{Output: "memory: no working directory set"}, nil
	}

	store, err := memory.NewStore(deps.Cwd)
	if err != nil {
		return Result{Output: fmt.Sprintf("memory: %v", err)}, nil
	}
	entries, err := store.List()
	if err != nil {
		return Result{Output: fmt.Sprintf("memory: %v", err)}, nil
	}
	if len(entries) == 0 {
		return Result{Output: "memory: no entries"}, nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Memory entries (%d):\n", len(entries))
	for _, name := range entries {
		fmt.Fprintf(&b, "  %s\n", name)
	}
	return Result{Output: b.String()}, nil
}
