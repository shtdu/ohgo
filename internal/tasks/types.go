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

// Copy returns a shallow clone of the record. Pointer fields (StartedAt,
// EndedAt, ReturnCode) are deep-copied so the caller can't mutate shared state.
func (r *Record) Copy() *Record {
	c := *r
	if r.StartedAt != nil {
		t := *r.StartedAt
		c.StartedAt = &t
	}
	if r.EndedAt != nil {
		t := *r.EndedAt
		c.EndedAt = &t
	}
	if r.ReturnCode != nil {
		v := *r.ReturnCode
		c.ReturnCode = &v
	}
	if r.Metadata != nil {
		c.Metadata = make(map[string]string, len(r.Metadata))
		for k, v := range r.Metadata {
			c.Metadata[k] = v
		}
	}
	return &c
}
