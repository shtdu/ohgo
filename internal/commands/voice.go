package commands

import "context"

// voiceCmd toggles voice mode (placeholder).
type voiceCmd struct{}

var _ Command = voiceCmd{}

func (voiceCmd) Name() string      { return "voice" }
func (voiceCmd) ShortHelp() string { return "Toggle voice mode" }

func (voiceCmd) Run(_ context.Context, _ string, _ *Deps) (Result, error) {
	return Result{Output: "voice: voice mode not yet implemented"}, nil
}

// VoiceCmd returns a new voice command.
func VoiceCmd() Command { return voiceCmd{} }
