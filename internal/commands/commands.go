// Package commands implements slash commands (/help, /commit, /plan, etc.).
package commands

import (
	"context"
)

// Command represents a slash command handler.
type Command interface {
	// Name returns the command name (e.g. "help", "commit").
	Name() string

	// Run executes the command.
	Run(ctx context.Context, args string) error
}

// Registry manages available slash commands.
type Registry struct {
	commands map[string]Command
}

// NewRegistry creates a new command registry.
func NewRegistry() *Registry {
	return &Registry{commands: make(map[string]Command)}
}

// Register adds a command.
func (r *Registry) Register(c Command) {
	r.commands[c.Name()] = c
}

// Get retrieves a command by name.
func (r *Registry) Get(name string) Command {
	return r.commands[name]
}
