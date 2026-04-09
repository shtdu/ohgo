package bridge

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBridge is a test bridge implementation.
type mockBridge struct {
	name      string
	connected bool
	connectErr error
	closeErr  error
}

func (m *mockBridge) Name() string { return m.name }
func (m *mockBridge) Connect(_ context.Context) error {
	if m.connectErr != nil {
		return m.connectErr
	}
	m.connected = true
	return nil
}
func (m *mockBridge) Close() error {
	if m.closeErr != nil {
		return m.closeErr
	}
	m.connected = false
	return nil
}

func TestManager_RegisterAndConnect(t *testing.T) {
	m := NewManager()
	b := &mockBridge{name: "test"}
	m.Register(b)

	err := m.ConnectAll(context.Background())
	require.NoError(t, err)
	assert.True(t, b.connected)
}

func TestManager_CloseAll(t *testing.T) {
	m := NewManager()
	b1 := &mockBridge{name: "a"}
	b2 := &mockBridge{name: "b"}
	m.Register(b1)
	m.Register(b2)

	require.NoError(t, m.ConnectAll(context.Background()))
	require.NoError(t, m.CloseAll())
	assert.False(t, b1.connected)
	assert.False(t, b2.connected)
}

func TestManager_Get(t *testing.T) {
	m := NewManager()
	b := &mockBridge{name: "found"}
	m.Register(b)

	got, ok := m.Get("found")
	assert.True(t, ok)
	assert.Equal(t, b, got)

	_, ok = m.Get("missing")
	assert.False(t, ok)
}

func TestManager_Status(t *testing.T) {
	m := NewManager()
	m.Register(&mockBridge{name: "a"})
	m.Register(&mockBridge{name: "b"})

	statuses := m.Status()
	assert.Len(t, statuses, 2)
	names := map[string]bool{}
	for _, s := range statuses {
		names[s.Name] = true
	}
	assert.True(t, names["a"])
	assert.True(t, names["b"])
}

func TestSessionRunner_StartStop(t *testing.T) {
	m := NewManager()
	b := &mockBridge{name: "test"}
	m.Register(b)

	runner := NewSessionRunner(m)

	session, err := runner.Start(context.Background(), "test")
	require.NoError(t, err)
	assert.Equal(t, "test-session", session.ID)
	assert.True(t, b.connected)

	err = runner.Stop("test-session")
	require.NoError(t, err)
	assert.False(t, b.connected)
}

func TestSessionRunner_BridgeNotFound(t *testing.T) {
	m := NewManager()
	runner := NewSessionRunner(m)

	_, err := runner.Start(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSessionRunner_SessionNotFound(t *testing.T) {
	m := NewManager()
	runner := NewSessionRunner(m)

	err := runner.Stop("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestNewClaudeCLI(t *testing.T) {
	c := NewClaudeCLI()
	assert.Equal(t, "claude", c.Name())
	assert.False(t, c.IsConnected())
}

func TestNewCodexBridge(t *testing.T) {
	c := NewCodexBridge()
	assert.Equal(t, "codex", c.Name())
	assert.False(t, c.IsConnected())
}

// Compile-time interface checks.
func TestBridgeInterface(t *testing.T) {
	var _ Bridge = (*ClaudeCLI)(nil)
	var _ Bridge = (*CodexBridge)(nil)
	var _ Bridge = (*mockBridge)(nil)
}

func TestSecretManager_GenerateAndValidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	sm := NewSecretManager(path)

	secret, err := sm.Generate("test-bridge")
	require.NoError(t, err)
	assert.NotEmpty(t, secret.ID)
	assert.NotEmpty(t, secret.Value)
	assert.Equal(t, "test-bridge", secret.BridgeName)

	// Validate correct secret.
	valid, err := sm.Validate(secret.ID, secret.Value)
	require.NoError(t, err)
	assert.True(t, valid)

	// Validate wrong value.
	valid, err = sm.Validate(secret.ID, "wrong-value")
	require.NoError(t, err)
	assert.False(t, valid)

	// Validate nonexistent ID.
	valid, err = sm.Validate("nonexistent", "any")
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestSecretManager_Rotate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	sm := NewSecretManager(path)

	secret, err := sm.Generate("test")
	require.NoError(t, err)
	oldValue := secret.Value

	rotated, err := sm.Rotate(secret.ID)
	require.NoError(t, err)
	assert.NotEqual(t, oldValue, rotated.Value)
	assert.Equal(t, secret.ID, rotated.ID)

	// Old value should no longer work.
	valid, _ := sm.Validate(secret.ID, oldValue)
	assert.False(t, valid)

	// New value should work.
	valid, _ = sm.Validate(secret.ID, rotated.Value)
	assert.True(t, valid)
}

func TestSecretManager_Revoke(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	sm := NewSecretManager(path)

	secret, err := sm.Generate("test")
	require.NoError(t, err)

	require.NoError(t, sm.Revoke(secret.ID))

	valid, _ := sm.Validate(secret.ID, secret.Value)
	assert.False(t, valid)
}

func TestSecretManager_RevokeNotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	sm := NewSecretManager(path)

	err := sm.Revoke("nonexistent")
	require.Error(t, err)
}

func TestSecretManager_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.json")
	sm := NewSecretManager(path)

	_, err := sm.Generate("test")
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}
