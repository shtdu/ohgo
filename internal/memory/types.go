package memory

import "time"

// Header holds metadata for one memory file.
type Header struct {
	Path        string
	Title       string
	Description string
	ModifiedAt  time.Time
	MemoryType  string
	BodyPreview string
}
