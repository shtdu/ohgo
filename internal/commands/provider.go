package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// providerCmd shows or switches the active provider profile.
type providerCmd struct{}

var _ Command = providerCmd{}

func (providerCmd) Name() string      { return "provider" }
func (providerCmd) ShortHelp() string { return "Show or switch the active provider profile" }

func (providerCmd) Run(_ context.Context, args string, deps *Deps) (Result, error) {
	arg := strings.TrimSpace(args)

	if arg == "" {
		// Show current profile details
		name, profile := deps.Config.ResolveProfile("")
		return Result{Output: formatKV(
			"Profile:", name,
			"Provider:", profile.Provider,
			"API format:", profile.APIFormat,
			"Default model:", profile.DefaultModel,
			"Current model:", profile.ResolvedModel(),
		)}, nil
	}

	// Switch profile
	profiles := deps.Config.MergedProfiles()
	if _, ok := profiles[arg]; !ok {
		// List available profiles
		var available []string
		for k := range profiles {
			available = append(available, k)
		}
		sort.Strings(available)
		return Result{}, fmt.Errorf("unknown profile %q; available: %s", arg, strings.Join(available, ", "))
	}

	prev := deps.Config.ActiveProfile
	deps.Config.ActiveProfile = arg

	// Also update the engine model to the profile's default
	profile := profiles[arg]
	deps.Engine.SetModel(profile.ResolvedModel())

	return Result{Output: fmt.Sprintf("Provider changed: %s -> %s", prev, arg)}, nil
}
