package gitserver

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type ObjectStorage interface {
	// GetStore gets an object store given the id where the id is the namespace
	// for storage
	GetStore(string) storer.Storer
}

func NewMemObjectStorage() *MemObjectStorage {
	return &MemObjectStorage{m: map[string]storer.Storer{}}
}

// MemObjectStorage manages objects stores by id i.e. repo
type MemObjectStorage struct {
	mu sync.Mutex
	m  map[string]storer.Storer
}

// GetStore for the given id.  Create one if it does not exist
func (mos *MemObjectStorage) GetStore(id string) storer.Storer {
	mos.mu.Lock()
	defer mos.mu.Unlock()

	if v, ok := mos.m[id]; ok {
		return v
	}

	mem := memory.NewStorage()
	mos.m[id] = mem
	return mem
}

// FilesystemObjectStorage manages objects stores by id i.e. repo
type FilesystemObjectStorage struct {
	mu      sync.Mutex
	datadir string
	m       map[string]storer.Storer
}

func NewFilesystemObjectStorage(dir string) *FilesystemObjectStorage {
	return &FilesystemObjectStorage{
		datadir: dir,
		m:       map[string]storer.Storer{},
	}
}

// GetStore for the given id.  Create one if it does not exist
func (mos *FilesystemObjectStorage) GetStore(id string) storer.Storer {
	mos.mu.Lock()
	defer mos.mu.Unlock()

	if v, ok := mos.m[id]; ok {
		return v
	}

	dir := filepath.Join(mos.datadir, id)
	os.MkdirAll(dir, 0755)
	fs := filesystem.NewStorage(osfs.New(dir), cache.NewObjectLRUDefault())
	mos.m[id] = fs
	return fs
}
