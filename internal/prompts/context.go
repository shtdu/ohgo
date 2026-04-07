package prompts

import (
	"context"
	"fmt"
	"os"
)

// Assembler builds the complete system prompt from base prompt, environment info,
// and CLAUDE.md instructions.
type Assembler struct {
	customPrompt string
	cwd          string
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
	if claudeMdContent == nil {
		return systemPrompt, nil
	}

	return systemPrompt + "\n\n" + *claudeMdContent, nil
}
