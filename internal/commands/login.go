package commands

import (
	"context"
	"fmt"
	"strings"
)

type loginCmd struct{}

var _ Command = loginCmd{}

func (loginCmd) Name() string      { return "login" }
func (loginCmd) ShortHelp() string { return "show or manage API authentication" }

func (loginCmd) Run(_ context.Context, args string, deps *Deps) (Result, error) {
	if deps.Config == nil {
		return Result{Output: "login: no configuration available"}, nil
	}

	args = strings.TrimSpace(args)

	if args == "logout" {
		deps.Config.APIKey = ""
		return Result{Output: "login: API key cleared"}, nil
	}

	if strings.HasPrefix(args, "set ") {
		key := strings.TrimSpace(strings.TrimPrefix(args, "set "))
		if key == "" {
			return Result{Output: "login set: missing API key"}, nil
		}
		deps.Config.APIKey = key
		return Result{Output: "login: API key updated"}, nil
	}

	// Show current status.
	if deps.Config.APIKey != "" {
		masked := maskKey(deps.Config.APIKey)
		return Result{Output: fmt.Sprintf("login: authenticated (key: %s)", masked)}, nil
	}
	return Result{Output: "login: not authenticated (use /login set <key> to set API key)"}, nil
}

// maskKey returns a masked version of an API key for display.
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
