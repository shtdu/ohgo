package commands

import (
	"context"
)

type onboardingCmd struct{}

var _ Command = onboardingCmd{}

func (onboardingCmd) Name() string      { return "onboarding" }
func (onboardingCmd) ShortHelp() string { return "show quickstart guide" }

func (onboardingCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: `OpenHarness Quickstart Guide
============================

1. Set your API key:
   /login set <your-api-key>

2. Verify authentication:
   /login

3. Start chatting:
   Type your prompt and press Enter.

4. Useful commands:
   /skills        - List loaded skills
   /tasks         - List background tasks
   /plugin        - Manage plugins
   /memory        - View memory entries
   /help          - Show all commands

5. Configuration:
   Settings are stored in ~/.openharness/settings.json
   Project-specific config goes in .openharness/settings.json

6. Learn more:
   https://github.com/HKUDS/OpenHarness
`}, nil
}
