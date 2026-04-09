package bridge

import (
	"context"
	"sync"
)

// BridgeStatus represents the current state of a bridge.
type BridgeStatus struct {
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
}

// Manager manages bridge lifecycle for subscription-backed providers.
type Manager struct {
	bridges map[string]Bridge
	mu      sync.RWMutex
}

// NewManager creates a new bridge manager.
func NewManager() *Manager {
	return &Manager{
		bridges: make(map[string]Bridge),
	}
}

// Register adds a bridge to the manager.
func (m *Manager) Register(b Bridge) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bridges[b.Name()] = b
}

// ConnectAll establishes connections for all registered bridges.
func (m *Manager) ConnectAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, b := range m.bridges {
		if err := b.Connect(ctx); err != nil {
			return err
		}
	}
	return nil
}

// CloseAll gracefully shuts down all bridges.
func (m *Manager) CloseAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var firstErr error
	for _, b := range m.bridges {
		if err := b.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Status returns the status of all bridges.
func (m *Manager) Status() []BridgeStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]BridgeStatus, 0, len(m.bridges))
	for _, b := range m.bridges {
		statuses = append(statuses, BridgeStatus{
			Name: b.Name(),
			// Connected state is tracked by individual bridge implementations.
		})
	}
	return statuses
}

// Get returns a bridge by name.
func (m *Manager) Get(name string) (Bridge, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.bridges[name]
	return b, ok
}
