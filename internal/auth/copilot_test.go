package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestFlow(deviceCodeServer, accessTokenServer *httptest.Server) *CopilotDeviceFlow {
	f := NewCopilotDeviceFlow()
	f.DeviceCodeURL = deviceCodeServer.URL
	f.AccessTokenURL = accessTokenServer.URL
	f.httpClient = deviceCodeServer.Client()
	return f
}

func TestCopilotDeviceFlow_Success(t *testing.T) {
	codeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"device_code":"dc_123","user_code":"ABCD-1234","verification_uri":"https://github.com/login/device","interval":1,"expires_in":900}`)
	}))
	defer codeServer.Close()

	var pollCount int32
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&pollCount, 1)
		w.Header().Set("Content-Type", "application/json")
		if count < 3 {
			fmt.Fprintf(w, `{"error":"authorization_pending"}`)
		} else {
			fmt.Fprintf(w, `{"access_token":"ghu_test_token_123"}`)
		}
	}))
	defer tokenServer.Close()

	f := newTestFlow(codeServer, tokenServer)
	cred, err := f.Authenticate(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "copilot", cred.Provider)
	assert.Equal(t, "oauth_token", cred.Kind)
	assert.Equal(t, "ghu_test_token_123", cred.Value)
	assert.True(t, atomic.LoadInt32(&pollCount) >= 3)
}

func TestCopilotDeviceFlow_SlowDown(t *testing.T) {
	codeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"device_code":"dc_123","user_code":"ABCD-1234","verification_uri":"https://github.com/login/device","interval":1,"expires_in":900}`)
	}))
	defer codeServer.Close()

	var pollCount int32
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&pollCount, 1)
		w.Header().Set("Content-Type", "application/json")
		if count == 1 {
			fmt.Fprintf(w, `{"error":"slow_down"}`)
		} else {
			fmt.Fprintf(w, `{"access_token":"ghu_after_slowdown"}`)
		}
	}))
	defer tokenServer.Close()

	f := newTestFlow(codeServer, tokenServer)
	cred, err := f.Authenticate(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "ghu_after_slowdown", cred.Value)
}

func TestCopilotDeviceFlow_ExpiredDeviceCode(t *testing.T) {
	codeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"device_code":"dc_123","user_code":"ABCD-1234","verification_uri":"https://github.com/login/device","interval":1,"expires_in":1}`)
	}))
	defer codeServer.Close()

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"error":"authorization_pending"}`)
	}))
	defer tokenServer.Close()

	f := newTestFlow(codeServer, tokenServer)
	_, err := f.Authenticate(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestCopilotDeviceFlow_ContextCancel(t *testing.T) {
	codeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"device_code":"dc_123","user_code":"ABCD-1234","verification_uri":"https://github.com/login/device","interval":1,"expires_in":900}`)
	}))
	defer codeServer.Close()

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"error":"authorization_pending"}`)
	}))
	defer tokenServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	f := newTestFlow(codeServer, tokenServer)
	_, err := f.Authenticate(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCopilotDeviceFlow_DeviceCodeRequestFailure(t *testing.T) {
	codeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"invalid_client"}`)
	}))
	defer codeServer.Close()

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer tokenServer.Close()

	f := newTestFlow(codeServer, tokenServer)
	_, err := f.Authenticate(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "device code request failed")
}

func TestCopilotDeviceFlow_AccessTokenError(t *testing.T) {
	codeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"device_code":"dc_123","user_code":"ABCD-1234","verification_uri":"https://github.com/login/device","interval":1,"expires_in":900}`)
	}))
	defer codeServer.Close()

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"error":"expired_token","error_description":"device code has expired"}`)
	}))
	defer tokenServer.Close()

	f := newTestFlow(codeServer, tokenServer)
	_, err := f.Authenticate(context.Background())
	require.Error(t, err)
	// Should get the error_description from the token endpoint.
	assert.Contains(t, err.Error(), "device code has expired")
}
