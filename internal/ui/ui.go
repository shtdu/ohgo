// Package ui implements the terminal user interface.
package ui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
)

// UI manages terminal output and user interaction.
type UI struct {
	out io.Writer
	in  io.Reader
	err io.Writer
}

// New creates a new UI writing to the given writers.
func New(out io.Writer, in io.Reader) *UI {
	return &UI{out: out, in: in, err: os.Stderr}
}

// WithErrWriter sets the error writer (defaults to os.Stderr).
func (u *UI) WithErrWriter(w io.Writer) *UI {
	u.err = w
	return u
}

// Print writes a message to the terminal output.
func (u *UI) Print(msg string) {
	fmt.Fprint(u.out, msg)
}

// Printf writes a formatted message to the terminal output.
func (u *UI) Printf(format string, args ...any) {
	fmt.Fprintf(u.out, format, args...)
}

// Println writes a message with a newline to the terminal output.
func (u *UI) Println(msg string) {
	fmt.Fprintln(u.out, msg)
}

// PrintError writes an error message to the error writer.
func (u *UI) PrintError(msg string) {
	fmt.Fprintln(u.err, msg)
}

// Prompt displays a prompt and reads a line of user input.
func (u *UI) Prompt(ctx context.Context, prompt string) (string, error) {
	fmt.Fprint(u.out, prompt)
	lineCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(u.in)
		if !scanner.Scan() {
			errCh <- scanner.Err()
			return
		}
		lineCh <- scanner.Text()
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case line := <-lineCh:
		return line, nil
	case err := <-errCh:
		return "", err
	}
}
