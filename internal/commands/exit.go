package commands

import "context"

// exitCmd terminates the REPL session.
type exitCmd struct{}

var _ Command = exitCmd{}

func (exitCmd) Name() string        { return "exit" }
func (exitCmd) ShortHelp() string   { return "Exit the REPL session" }

func (exitCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: "Goodbye!", ShouldExit: true}, nil
}
