package commands

import "context"

// privacyCmd shows privacy settings (placeholder).
type privacyCmd struct{}

var _ Command = privacyCmd{}

func (privacyCmd) Name() string      { return "privacy-settings" }
func (privacyCmd) ShortHelp() string { return "Show privacy settings" }

func (privacyCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: "privacy-settings: default privacy settings active"}, nil
}

// PrivacyCmd returns a new privacy-settings command.
func PrivacyCmd() Command { return privacyCmd{} }
