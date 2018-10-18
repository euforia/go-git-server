package storage

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

// GitRepoStorage implements an interface to store objects for repos
type GitRepoStorage interface {
	// GetStore gets an object store given the id where the id is the namespace
	// for storage
	GetStore(string) storer.Storer
}

// MemGitRepoStorage manages objects stores by id i.e. repo
type MemGitRepoStorage struct {
	mu sync.Mutex
	m  map[string]storer.Storer
}

// NewMemGitRepoStorage returns a new instance of MemGitRepoStorage
func NewMemGitRepoStorage() *MemGitRepoStorage {
	return &MemGitRepoStorage{m: map[string]storer.Storer{}}
}

// GetStore for the given id.  Create one if it does not exist
func (mos *MemGitRepoStorage) GetStore(id string) storer.Storer {
	mos.mu.Lock()
	defer mos.mu.Unlock()

	if v, ok := mos.m[id]; ok {
		return v
	}

	mem := memory.NewStorage()
	mos.m[id] = mem
	return mem
}

// FilesystemGitRepoStorage manages objects stores by id i.e. repo
type FilesystemGitRepoStorage struct {
	mu      sync.Mutex
	datadir string
	m       map[string]storer.Storer
}

// NewFilesystemGitRepoStorage returns an new instance of FilesystemGitRepoStorage
func NewFilesystemGitRepoStorage(dir string) *FilesystemGitRepoStorage {
	return &FilesystemGitRepoStorage{
		datadir: dir,
		m:       map[string]storer.Storer{},
	}
}

// GetStore for the given id.  Create one if it does not exist
func (mos *FilesystemGitRepoStorage) GetStore(id string) storer.Storer {
	mos.mu.Lock()
	defer mos.mu.Unlock()

	if v, ok := mos.m[id]; ok {
		return v
	}

	dir := filepath.Join(mos.datadir, id)
	_, err := os.Stat(dir)
	if err != nil {
		return nil
	}

	fs := filesystem.NewStorage(osfs.New(dir), cache.NewObjectLRUDefault())
	mos.m[id] = fs
	return fs
}
