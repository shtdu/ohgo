package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shtdu/ohgo/internal/auth"
)

type loginCmd struct{}

var _ Command = loginCmd{}

func (loginCmd) Name() string      { return "login" }
func (loginCmd) ShortHelp() string { return "show or manage API authentication" }

func (loginCmd) Run(ctx context.Context, args string, deps *Deps) (Result, error) {
	args = strings.TrimSpace(args)

	if args == "logout" {
		return handleLogout(ctx, deps)
	}

	if strings.HasPrefix(args, "set ") {
		return handleLoginSet(ctx, args, deps)
	}

	// Show current status.
	return showLoginStatus(ctx, deps)
}

func handleLogout(ctx context.Context, deps *Deps) (Result, error) {
	if deps.AuthMgr == nil {
		if deps.Config != nil {
			deps.Config.APIKey = ""
		}
		return Result{Output: "login: API key cleared"}, nil
	}

	creds, err := deps.AuthMgr.List(ctx)
	if err != nil || len(creds) == 0 {
		return Result{Output: "login: no stored credentials to clear"}, nil
	}

	var anyErr bool
	for _, c := range creds {
		if err := deps.AuthMgr.Delete(ctx, c.Provider); err != nil {
			anyErr = true
		}
	}
	if anyErr {
		return Result{Output: "login: some credentials could not be cleared"}, nil
	}
	return Result{Output: "login: all stored credentials cleared"}, nil
}

func handleLoginSet(ctx context.Context, args string, deps *Deps) (Result, error) {
	parts := strings.Fields(strings.TrimPrefix(args, "set "))
	if len(parts) == 0 || parts[0] == "" {
		return Result{Output: "login set: missing provider and key\nusage: /login set <provider> <key>"}, nil
	}

	provider := parts[0]
	if len(parts) < 2 || parts[1] == "" {
		return Result{Output: fmt.Sprintf("login set: missing key for %s\nusage: /login set <provider> <key>", provider)}, nil
	}
	key := parts[1]

	if deps.AuthMgr != nil {
		cred := &auth.Credential{
			Provider:  provider,
			Kind:      "api_key",
			Value:     key,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		if err := deps.AuthMgr.Store(ctx, cred); err != nil {
			return Result{Output: fmt.Sprintf("login: failed to store credential: %v", err)}, nil
		}
	}

	// Also set on config for current session.
	if deps.Config != nil {
		deps.Config.APIKey = key
	}

	return Result{Output: fmt.Sprintf("login: credential stored for %s (%s)", provider, maskKey(key))}, nil
}

func showLoginStatus(ctx context.Context, deps *Deps) (Result, error) {
	var b strings.Builder

	// Show stored credentials.
	if deps.AuthMgr != nil {
		creds, err := deps.AuthMgr.List(ctx)
		if err == nil && len(creds) > 0 {
			b.WriteString("Stored Credentials:\n")
			for _, c := range creds {
				fmt.Fprintf(&b, "  %-12s %s (%s)\n", c.Provider, maskKey(c.Value), c.Kind)
			}
		}
	}

	// Show current session key.
	if deps.Config != nil && deps.Config.APIKey != "" {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "Session Key: %s\n", maskKey(deps.Config.APIKey))
	}

	if b.Len() == 0 {
		return Result{Output: "login: not authenticated\nuse /login set <provider> <key> to store a credential"}, nil
	}

	return Result{Output: b.String()}, nil
}

// maskKey returns a masked version of an API key for display.
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
