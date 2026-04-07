// Package prompts handles system prompt assembly and CLAUDE.md discovery.
package prompts

import (
	"context"
)

// Assembler builds the system prompt from various sources.
type Assembler struct{}

// NewAssembler creates a new prompt assembler.
func NewAssembler() *Assembler {
	return &Assembler{}
}

// Build assembles the full system prompt, including CLAUDE.md content and skills.
func (a *Assembler) Build(ctx context.Context) (string, error) {
	// TODO: implement prompt assembly
	// - Discover CLAUDE.md files (project, user global)
	// - Load active skills
	// - Compose final system prompt
	return "", nil
}
