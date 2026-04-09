package commands

import "context"

// clearCmd resets the conversation history.
type clearCmd struct{}

var _ Command = clearCmd{}

func (clearCmd) Name() string      { return "clear" }
func (clearCmd) ShortHelp() string { return "Clear the conversation history" }

func (clearCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	deps.Engine.Clear()
	return Result{Output: "Conversation cleared."}, nil
}
