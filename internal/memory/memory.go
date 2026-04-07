// Package memory implements persistent cross-session memory.
package memory

import (
	"context"
)

// Store provides persistent key-value storage across sessions.
type Store struct {
	dir string
}

// NewStore creates a memory store backed by the given directory.
func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// Save writes a memory entry.
func (s *Store) Save(ctx context.Context, key, value string) error {
	// TODO: implement persistent storage
	return nil
}

// Load reads a memory entry.
func (s *Store) Load(ctx context.Context, key string) (string, error) {
	// TODO: implement persistent storage
	return "", nil
}
