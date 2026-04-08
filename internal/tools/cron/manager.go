// Package cron provides an in-memory cron job manager.
package cron

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Job represents a scheduled cron job definition.
type Job struct {
	Name       string     `json:"name"`
	Schedule   string     `json:"schedule"`
	Command    string     `json:"command"`
	Cwd        string     `json:"cwd,omitempty"`
	Enabled    bool       `json:"enabled"`
	CreatedAt  time.Time  `json:"created_at"`
	NextRun    *time.Time `json:"next_run,omitempty"`
	LastRun    *time.Time `json:"last_run,omitempty"`
	LastStatus string     `json:"last_status,omitempty"`
}

// Manager manages in-memory cron job definitions.
type Manager struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

// NewManager creates an empty cron job manager.
func NewManager() *Manager {
	return &Manager{jobs: make(map[string]*Job)}
}

// ValidateSchedule checks whether a cron expression is valid.
func ValidateSchedule(expr string) error {
	_, err := cron.ParseStandard(expr)
	return err
}

// Create adds a new job. Returns an error if the schedule is invalid.
func (m *Manager) Create(job Job) error {
	if err := ValidateSchedule(job.Schedule); err != nil {
		return fmt.Errorf("invalid schedule %q: %w", job.Schedule, err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	job.CreatedAt = time.Now()
	m.jobs[job.Name] = &job
	return nil
}

// Delete removes a job by name. Returns false if not found.
func (m *Manager) Delete(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.jobs[name]; !ok {
		return false
	}
	delete(m.jobs, name)
	return true
}

// List returns a snapshot of all jobs.
func (m *Manager) List() []Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		out = append(out, *j)
	}
	return out
}

// Get retrieves a single job by name.
func (m *Manager) Get(name string) (*Job, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	j, ok := m.jobs[name]
	if !ok {
		return nil, false
	}
	cp := *j
	return &cp, true
}

// Toggle enables or disables a job. Returns false if not found.
func (m *Manager) Toggle(name string, enabled bool) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[name]
	if !ok {
		return false
	}
	j.Enabled = enabled
	return true
}
