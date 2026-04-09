// Package auth implements authentication flows (OAuth device flow, API key, etc.).
package auth

import "context"

// Provider authenticates with an LLM provider.
type Provider interface {
	// Authenticate runs the auth flow and returns an API key or token.
	Authenticate(ctx context.Context) (string, error)
}
