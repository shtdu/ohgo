package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shtdu/ohgo/internal/api"
)

// resumeCmd restores a tagged session.
type resumeCmd struct{}

var _ Command = resumeCmd{}

func (resumeCmd) Name() string     { return "resume" }
func (resumeCmd) ShortHelp() string { return "restore a tagged session" }

func (resumeCmd) Run(_ context.Context, args string, deps *Deps) (Result, error) {
	tag := strings.TrimSpace(args)
	if tag == "" {
		return Result{}, fmt.Errorf("resume: tag name required")
	}

	dir := sessionDir()
	path := filepath.Join(dir, tag+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Result{}, fmt.Errorf("resume: tag %q not found", tag)
		}
		return Result{}, fmt.Errorf("resume: read: %w", err)
	}

	var msgs []api.Message
	if err := json.Unmarshal(data, &msgs); err != nil {
		return Result{}, fmt.Errorf("resume: parse: %w", err)
	}

	deps.Engine.LoadMessages(msgs)

	return Result{
		Output: fmt.Sprintf("resume: restored %d messages from tag %q", len(msgs), tag),
	}, nil
}
