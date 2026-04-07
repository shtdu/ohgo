// Package auth implements authentication flows (OAuth device flow, API key, etc.).
package auth

import (
	"context"
)

// Provider authenticates with an LLM provider.
type Provider interface {
	// Authenticate runs the auth flow and returns an API key or token.
	Authenticate(ctx context.Context) (string, error)
}

// OAuthDeviceFlow implements the OAuth 2.0 device authorization grant.
type OAuthDeviceFlow struct {
	ClientID string
	Audience string
}

// Authenticate performs the device flow and returns an access token.
func (o *OAuthDeviceFlow) Authenticate(ctx context.Context) (string, error) {
	// TODO: implement OAuth device flow
	return "", nil
}
