package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	copilotClientID      = "Iv1.b507a08c87ecfe98"
	githubDeviceCodeURL  = "https://github.com/login/device/code"
	githubAccessTokenURL = "https://github.com/login/oauth/access_token"
)

// CopilotDeviceFlow implements OAuth 2.0 Device Authorization Grant for GitHub Copilot.
type CopilotDeviceFlow struct {
	ClientID string
}

// NewCopilotDeviceFlow creates a new Copilot device flow.
func NewCopilotDeviceFlow() *CopilotDeviceFlow {
	return &CopilotDeviceFlow{ClientID: copilotClientID}
}

// Name returns the flow name.
func (f *CopilotDeviceFlow) Name() string { return "device_code" }

// Authenticate performs the GitHub device code flow for Copilot access.
func (f *CopilotDeviceFlow) Authenticate(ctx context.Context) (*Credential, error) {
	// Step 1: Request device code.
	codeResp, err := f.requestDeviceCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("copilot auth: request device code: %w", err)
	}

	// Step 2: Display user code and verification URL.
	fmt.Fprintf(os.Stderr, "\nTo authenticate with GitHub Copilot:\n")
	fmt.Fprintf(os.Stderr, "  1. Open: %s\n", codeResp.VerificationURI)
	fmt.Fprintf(os.Stderr, "  2. Enter code: %s\n\n", codeResp.UserCode)

	// Step 3: Poll for access token.
	token, err := f.pollForToken(ctx, codeResp)
	if err != nil {
		return nil, fmt.Errorf("copilot auth: %w", err)
	}

	now := time.Now().Unix()
	return &Credential{
		Provider:  "copilot",
		Kind:      "oauth_token",
		Value:     token,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	Interval        int    `json:"interval"` // seconds between polls
	ExpiresIn       int    `json:"expires_in"`
}

func (f *CopilotDeviceFlow) requestDeviceCode(ctx context.Context) (*deviceCodeResponse, error) {
	data := url.Values{
		"client_id": {f.ClientID},
		"scope":     {"copilot"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubDeviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device code request failed (status %d): %s", resp.StatusCode, body)
	}

	var codeResp deviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&codeResp); err != nil {
		return nil, fmt.Errorf("decode device code response: %w", err)
	}

	return &codeResp, nil
}

func (f *CopilotDeviceFlow) pollForToken(ctx context.Context, codeResp *deviceCodeResponse) (string, error) {
	interval := time.Duration(codeResp.Interval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}
	deadline := time.After(time.Duration(codeResp.ExpiresIn) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-deadline:
			return "", fmt.Errorf("device code expired")
		case <-time.After(interval):
			token, err := f.checkToken(ctx, codeResp.DeviceCode)
			if err != nil {
				// "authorization_pending" means user hasn't authorized yet.
				if strings.Contains(err.Error(), "authorization_pending") {
					continue
				}
				if strings.Contains(err.Error(), "slow_down") {
					interval += 5 * time.Second
					continue
				}
				return "", err
			}
			return token, nil
		}
	}
}

func (f *CopilotDeviceFlow) checkToken(ctx context.Context, deviceCode string) (string, error) {
	data := url.Values{
		"client_id":   {f.ClientID},
		"device_code": {deviceCode},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubAccessTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	if token, ok := result["access_token"].(string); ok && token != "" {
		return token, nil
	}

	if errDesc, ok := result["error_description"].(string); ok && errDesc != "" {
		return "", fmt.Errorf("%s", errDesc)
	}

	if errCode, ok := result["error"].(string); ok {
		return "", fmt.Errorf("%s", errCode)
	}

	return "", fmt.Errorf("unexpected token response")
}
