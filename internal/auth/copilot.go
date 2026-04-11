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
	copilotClientID       = "Iv1.b507a08c87ecfe98"
	defaultDeviceCodeURL  = "https://github.com/login/device/code"
	defaultAccessTokenURL = "https://github.com/login/oauth/access_token"
)

// CopilotDeviceFlow implements OAuth 2.0 Device Authorization Grant for GitHub Copilot.
type CopilotDeviceFlow struct {
	ClientID       string
	DeviceCodeURL  string
	AccessTokenURL string
	httpClient     *http.Client
}

// NewCopilotDeviceFlow creates a new Copilot device flow.
func NewCopilotDeviceFlow() *CopilotDeviceFlow {
	return &CopilotDeviceFlow{
		ClientID:       copilotClientID,
		DeviceCodeURL:  defaultDeviceCodeURL,
		AccessTokenURL: defaultAccessTokenURL,
		httpClient:     http.DefaultClient,
	}
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.DeviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

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
	if interval < 1*time.Second {
		interval = 1 * time.Second
	}
	deadline := time.After(time.Duration(codeResp.ExpiresIn) * time.Second)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-deadline:
			return "", fmt.Errorf("device code expired")
		case <-ticker.C:
			token, err := f.checkToken(ctx, codeResp.DeviceCode)
			if err != nil {
				if strings.Contains(err.Error(), "authorization_pending") {
					continue
				}
				if strings.Contains(err.Error(), "slow_down") {
					interval += 5 * time.Second
					ticker.Reset(interval)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.AccessTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

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
