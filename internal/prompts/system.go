package prompts

import (
	"context"
	"fmt"
	"strings"
)

const baseSystemPrompt = `You are og (OpenHarness Go), an open-source AI-powered coding assistant that runs in the terminal. You help users with software engineering tasks.

## Core Capabilities
- Read, write, and edit files
- Execute shell commands
- Search code and files
- Web search and fetch
- Multi-turn conversations with tool use

## Key Principles
- Be concise and direct
- Prefer the simplest approach that works
- Verify before asserting — read files before suggesting changes
- Use tools proactively to gather information
- Ask for clarification when requirements are ambiguous
- Never commit code without the user's explicit request

## Tool Usage
- Use tools to accomplish tasks — read files, run commands, search code
- Each tool call should have a clear purpose
- Prefer reading before writing — understand existing code first
- Break complex tasks into smaller steps

## Safety
- Never run destructive commands without explicit user approval
- Validate inputs at system boundaries
- Handle errors explicitly — don't swallow them
- Respect .gitignore and don't access sensitive files like .env`

// BuildSystemPrompt assembles the system prompt.
// If customPrompt is non-empty, it replaces the base prompt.
// If env is nil, it is auto-detected using DetectEnvironment.
func BuildSystemPrompt(ctx context.Context, customPrompt string, env *EnvironmentInfo) (string, error) {
	if env == nil {
		detected, err := DetectEnvironment(ctx, "")
		if err != nil {
			return "", fmt.Errorf("detect environment: %w", err)
		}
		env = &detected
	}

	prompt := customPrompt
	if prompt == "" {
		prompt = baseSystemPrompt
	}

	var envSection strings.Builder
	envSection.WriteString("\n\n## Environment\n")
	fmt.Fprintf(&envSection, "- OS: %s", env.OSName)
	if env.OSVersion != "" {
		fmt.Fprintf(&envSection, " %s", env.OSVersion)
	}
	fmt.Fprintf(&envSection, "\n- Architecture: %s", env.Architecture)
	fmt.Fprintf(&envSection, "\n- Shell: %s", env.Shell)
	fmt.Fprintf(&envSection, "\n- Working directory: %s", env.WorkingDir)
	fmt.Fprintf(&envSection, "\n- Date: %s", env.Date)
	fmt.Fprintf(&envSection, "\n- Go version: %s", env.GoVersion)
	if env.IsGitRepo {
		fmt.Fprintf(&envSection, "\n- Git branch: %s", env.GitBranch)
	}

	return prompt + envSection.String(), nil
}
