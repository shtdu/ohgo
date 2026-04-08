package tasks

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/shtdu/ohgo/internal/config"
)

// defaultMaxOutput is the default maximum bytes returned by ReadOutput.
const defaultMaxOutput = 12000

// Manager handles background task creation, monitoring, and lifecycle.
type Manager struct {
	mu          sync.RWMutex
	tasks       map[string]*Record
	processes   map[string]*os.Process
	cancelFuncs map[string]context.CancelFunc
	outputMu    map[string]*sync.Mutex
}

// NewManager creates a new task manager.
func NewManager() *Manager {
	return &Manager{
		tasks:       make(map[string]*Record),
		processes:   make(map[string]*os.Process),
		cancelFuncs: make(map[string]context.CancelFunc),
		outputMu:    make(map[string]*sync.Mutex),
	}
}

// taskID generates a unique task ID with a type prefix.
func taskID(taskType Type) string {
	prefixes := map[Type]string{
		TypeLocalBash:  "b",
		TypeLocalAgent: "a",
	}
	prefix := prefixes[taskType]

	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s%x", prefix, b)
}

// CreateShell starts a local bash command as a background task.
func (m *Manager) CreateShell(_ context.Context, command, description, cwd string) (*Record, error) {
	id := taskID(TypeLocalBash)

	tasksDir, err := config.TasksDir()
	if err != nil {
		return nil, fmt.Errorf("cannot get tasks directory: %w", err)
	}
	outputPath := filepath.Join(tasksDir, id+".log")

	now := time.Now()
	rec := &Record{
		ID:          id,
		Type:        TypeLocalBash,
		Status:      StatusRunning,
		Description: description,
		Cwd:         cwd,
		OutputFile:  outputPath,
		Command:     command,
		CreatedAt:   now,
		StartedAt:   &now,
		Metadata:    make(map[string]string),
	}

	// Create a cancellable context for this task.
	taskCtx, cancel := context.WithCancel(context.Background())

	m.mu.Lock()
	m.tasks[id] = rec
	m.cancelFuncs[id] = cancel
	m.outputMu[id] = &sync.Mutex{}
	m.mu.Unlock()

	// Open the output file for streaming. Use a mutex-protected writer
	// so ReadOutput can safely read while the process is still writing.
	outputFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		cancel()
		m.mu.Lock()
		delete(m.tasks, id)
		delete(m.cancelFuncs, id)
		delete(m.outputMu, id)
		m.mu.Unlock()
		return nil, fmt.Errorf("cannot create output file: %w", err)
	}

	// Wrap the file in a synchronized writer so concurrent ReadOutput calls are safe.
	syncWriter := &syncFileWriter{file: outputFile, mu: m.outputMu[id]}

	cmd := exec.CommandContext(taskCtx, "sh", "-c", command)
	cmd.Dir = cwd
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = syncWriter
	cmd.Stderr = syncWriter

	if err := cmd.Start(); err != nil {
		outputFile.Close()
		cancel()
		m.mu.Lock()
		delete(m.tasks, id)
		delete(m.cancelFuncs, id)
		delete(m.outputMu, id)
		m.mu.Unlock()
		return nil, fmt.Errorf("cannot start command: %w", err)
	}

	m.mu.Lock()
	m.processes[id] = cmd.Process
	m.mu.Unlock()

	// Goroutine to monitor process completion.
	go func() {
		waitErr := cmd.Wait()

		// Close the output file now that the process is done writing.
		outputFile.Close()

		m.mu.Lock()
		defer m.mu.Unlock()

		r, ok := m.tasks[id]
		if !ok || r.Status == StatusKilled {
			return
		}
		if waitErr != nil {
			r.Status = StatusFailed
		} else {
			r.Status = StatusCompleted
		}
		endNow := time.Now()
		r.EndedAt = &endNow
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			rc := exitErr.ExitCode()
			r.ReturnCode = &rc
		}
		delete(m.processes, id)
		delete(m.outputMu, id)
	}()

	return rec, nil
}

// syncFileWriter wraps an os.File with a mutex for safe concurrent access.
type syncFileWriter struct {
	file *os.File
	mu   *sync.Mutex
}

func (w *syncFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Write(p)
}

// Get retrieves a copy of a task record by ID.
// Returns a copy to avoid data races with background goroutines.
func (m *Manager) Get(id string) (*Record, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rec, ok := m.tasks[id]
	if !ok {
		return nil, false
	}
	return rec.Copy(), true
}

// List returns copies of task records, optionally filtered by status.
// Results are sorted by CreatedAt descending (newest first).
func (m *Manager) List(status Status) []*Record {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Record
	for _, rec := range m.tasks {
		if status != "" && rec.Status != status {
			continue
		}
		result = append(result, rec.Copy())
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result
}

// Update modifies a task's description and metadata.
func (m *Manager) Update(id string, description string, progress int, statusNote string) (*Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task %s not found", id)
	}
	if description != "" {
		rec.Description = description
	}
	if statusNote != "" {
		if rec.Metadata == nil {
			rec.Metadata = make(map[string]string)
		}
		rec.Metadata["statusNote"] = statusNote
	}
	if progress >= 0 {
		if rec.Metadata == nil {
			rec.Metadata = make(map[string]string)
		}
		rec.Metadata["progress"] = fmt.Sprintf("%d", progress)
	}
	return rec.Copy(), nil
}

// Stop terminates a running task by sending SIGTERM, then SIGKILL after a grace period.
func (m *Manager) Stop(_ context.Context, id string) error {
	m.mu.Lock()
	proc, ok := m.processes[id]
	if !ok {
		m.mu.Unlock()
		// Task may already be completed or not exist.
		rec, found := m.tasks[id]
		if found && (rec.Status == StatusCompleted || rec.Status == StatusFailed || rec.Status == StatusKilled) {
			return nil
		}
		return fmt.Errorf("task %s not found or not running", id)
	}
	m.mu.Unlock()

	// Send SIGTERM to the process group.
	if err := syscall.Kill(-proc.Pid, syscall.SIGTERM); err != nil {
		// Process may have already exited.
		_ = syscall.Kill(-proc.Pid, syscall.SIGKILL)
	}

	// Wait up to 3 seconds for the process to exit.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		// Check if process still exists by sending signal 0.
		if err := syscall.Kill(-proc.Pid, 0); err != nil {
			// Process has exited.
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Force kill if still running.
	_ = syscall.Kill(-proc.Pid, syscall.SIGKILL)

	// Cancel the task context.
	if cancel, ok := m.cancelFuncs[id]; ok {
		cancel()
	}

	m.mu.Lock()
	rec, ok := m.tasks[id]
	if ok {
		rec.Status = StatusKilled
		endNow := time.Now()
		rec.EndedAt = &endNow
	}
	delete(m.processes, id)
	delete(m.cancelFuncs, id)
	delete(m.outputMu, id)
	m.mu.Unlock()

	return nil
}

// ReadOutput returns the task's captured output.
// If the output exceeds maxBytes, only the last maxBytes are returned.
// If maxBytes <= 0, defaultMaxOutput is used.
func (m *Manager) ReadOutput(id string, maxBytes int) (string, error) {
	m.mu.RLock()
	rec, ok := m.tasks[id]
	outputMu := m.outputMu[id]
	m.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("task %s not found", id)
	}

	if outputMu != nil {
		outputMu.Lock()
	}
	data, err := os.ReadFile(rec.OutputFile)
	if outputMu != nil {
		outputMu.Unlock()
	}
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("cannot read output for task %s: %w", id, err)
	}

	if maxBytes <= 0 {
		maxBytes = defaultMaxOutput
	}

	output := string(data)
	if len(output) > maxBytes {
		output = output[len(output)-maxBytes:]
	}
	return output, nil
}
