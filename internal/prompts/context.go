package prompts

import (
	"context"
	"fmt"
	"os"
)

// MemoryLoader abstracts memory loading for prompt injection.
// This keeps prompts decoupled from the concrete memory package.
type MemoryLoader interface {
	LoadPrompt(maxLines int) (string, error)
}

// Assembler builds the complete system prompt from base prompt, environment info,
// CLAUDE.md instructions, and memory content.
type Assembler struct {
	customPrompt string
	cwd          string
	memory       MemoryLoader
}

// NewAssembler creates a prompt assembler for the given working directory.
func NewAssembler(cwd string) *Assembler {
	return &Assembler{cwd: cwd}
}

// WithCustomPrompt sets a custom base prompt (overrides built-in).
func (a *Assembler) WithCustomPrompt(prompt string) *Assembler {
	a.customPrompt = prompt
	return a
}

// WithMemoryStore sets the memory store for prompt injection.
func (a *Assembler) WithMemoryStore(m MemoryLoader) *Assembler {
	a.memory = m
	return a
}

// Build assembles the full system prompt.
func (a *Assembler) Build(ctx context.Context) (string, error) {
	cwd := a.cwd
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	// Detect environment with our cwd
	env, err := DetectEnvironment(ctx, cwd)
	if err != nil {
		return "", fmt.Errorf("detect environment: %w", err)
	}

	systemPrompt, err := BuildSystemPrompt(ctx, a.customPrompt, &env)
	if err != nil {
		return "", fmt.Errorf("build system prompt: %w", err)
	}

	files, err := DiscoverCLAUDEmd(ctx, cwd)
	if err != nil {
		return systemPrompt, nil // CLAUDE.md failure is non-fatal
	}

	claudeMdContent := MergeCLAUDEmd(files, 12000)
	if claudeMdContent != nil {
		systemPrompt += "\n\n" + *claudeMdContent
	}

	// Inject memory content if a store is configured.
	if a.memory != nil {
		memContent, err := a.memory.LoadPrompt(200)
		if err == nil && memContent != "" {
			systemPrompt += "\n\n" + memContent
		}
	}

	return systemPrompt, nil
}
