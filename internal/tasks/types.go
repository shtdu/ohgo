package tasks

import "time"

// Type represents the kind of background task.
type Type string

// Status represents the current state of a task.
type Status string

const (
	TypeLocalBash  Type = "local_bash"
	TypeLocalAgent Type = "local_agent"

	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusKilled    Status = "killed"
)

// Record holds the persistent state of a background task.
type Record struct {
	ID          string
	Type        Type
	Status      Status
	Description string
	Cwd         string
	OutputFile  string
	Command     string
	Prompt      string
	CreatedAt   time.Time
	StartedAt   *time.Time
	EndedAt     *time.Time
	ReturnCode  *int
	Metadata    map[string]string
}
