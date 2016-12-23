package gitserver

import (
	"sync"

	"github.com/src-d/go-git/storage/memory"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

type ObjectStorage interface {
	// GetStore gets an object store given the id where the id is the namespace
	// for storage
	GetStore(string) storer.EncodedObjectStorer
}

func NewMemObjectStorage() *MemObjectStorage {
	return &MemObjectStorage{m: map[string]storer.EncodedObjectStorer{}}
}

// MemObjectStorage manages objects stores by id i.e. repo
type MemObjectStorage struct {
	mu sync.Mutex
	m  map[string]storer.EncodedObjectStorer
}

// GetStore for the given id.  Create one if it does not exist
func (mos *MemObjectStorage) GetStore(id string) storer.EncodedObjectStorer {
	mos.mu.Lock()
	defer mos.mu.Unlock()

	if v, ok := mos.m[id]; ok {
		return v
	}

	mem := memory.NewStorage()
	mos.m[id] = mem
	return mem
}
