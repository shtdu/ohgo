package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shtdu/ohgo/internal/config"
)

type feedbackCmd struct{}

var _ Command = feedbackCmd{}

func (feedbackCmd) Name() string      { return "feedback" }
func (feedbackCmd) ShortHelp() string { return "save feedback" }

func (feedbackCmd) Run(_ context.Context, args string, _ *Deps) (Result, error) {
	args = strings.TrimSpace(args)
	if args == "" {
		return Result{Output: "feedback: please provide feedback text (e.g. /feedback your text here)"}, nil
	}

	dir, err := config.ConfigDir()
	if err != nil {
		return Result{}, fmt.Errorf("feedback: cannot determine config dir: %w", err)
	}

	feedbackPath := filepath.Join(dir, "feedback.txt")

	entry := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), args)

	f, err := os.OpenFile(feedbackPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return Result{}, fmt.Errorf("feedback: cannot write file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString(entry); err != nil {
		return Result{}, fmt.Errorf("feedback: write failed: %w", err)
	}

	return Result{Output: "feedback: thank you, your feedback has been saved"}, nil
}
