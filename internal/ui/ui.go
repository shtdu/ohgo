// Package ui implements the terminal user interface.
package ui

import (
	"context"
	"io"
)

// UI manages terminal output and user interaction.
type UI struct {
	out io.Writer
	in  io.Reader
}

// New creates a new UI writing to the given writer.
func New(out io.Writer, in io.Reader) *UI {
	return &UI{out: out, in: in}
}

// Print writes a message to the terminal.
func (u *UI) Print(msg string) {
	// TODO: implement terminal output with formatting
}

// Prompt displays a prompt and reads user input.
func (u *UI) Prompt(ctx context.Context, prompt string) (string, error) {
	// TODO: implement interactive prompt
	return "", nil
}
