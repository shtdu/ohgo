package commands

import (
	"context"
	"fmt"
	"strings"
)

// modelCmd shows or switches the current model.
type modelCmd struct{}

var _ Command = modelCmd{}

func (modelCmd) Name() string      { return "model" }
func (modelCmd) ShortHelp() string { return "Show or switch the current model" }

func (modelCmd) Run(_ context.Context, args string, deps *Deps) (Result, error) {
	arg := strings.TrimSpace(args)
	if arg == "" {
		return Result{Output: fmt.Sprintf("Current model: %s", deps.Engine.Model())}, nil
	}
	prev := deps.Engine.Model()
	deps.Engine.SetModel(arg)
	return Result{Output: fmt.Sprintf("Model changed: %s -> %s", prev, arg)}, nil
}
