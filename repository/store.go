package repository

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("exists")
)

// RepositoryStore is the repository storage interface that should be implemented.
type RepositoryStore interface {
	GetRepo(string) (*Repository, error)
	CreateRepo(*Repository) error
	UpdateRepo(*Repository) error
	RemoveRepo(string) error
}

// MemRepoStore is a memory based repo datastore
type MemRepoStore struct {
	mu sync.Mutex
	m  map[string]*Repository
}

// NewMemRepoStore instantiates a new repo store
func NewMemRepoStore() *MemRepoStore {
	return &MemRepoStore{m: map[string]*Repository{}}
}

// GetRepo with the given id
func (mrs *MemRepoStore) GetRepo(id string) (*Repository, error) {
	if v, ok := mrs.m[id]; ok {
		return v, nil
	}
	return nil, ErrNotFound
}

// CreateRepo with the given repo data
func (mrs *MemRepoStore) CreateRepo(repo *Repository) error {
	mrs.mu.Lock()
	defer mrs.mu.Unlock()

	if _, ok := mrs.m[repo.ID]; ok {
		return ErrExists
	}

	mrs.m[repo.ID] = repo
	return nil
}

// UpdateRepo with the given data
func (mrs *MemRepoStore) UpdateRepo(repo *Repository) error {
	mrs.mu.Lock()
	defer mrs.mu.Unlock()

	if _, ok := mrs.m[repo.ID]; ok {
		mrs.m[repo.ID] = repo
		return nil
	}
	return ErrNotFound
}

// RemoveRepo with the given id
func (mrs *MemRepoStore) RemoveRepo(id string) error {
	mrs.mu.Lock()
	defer mrs.mu.Unlock()

	if _, ok := mrs.m[id]; ok {
		delete(mrs.m, id)
		return nil
	}
	return ErrNotFound
}

// FilesystemRepoStore is a file based repo datastore
type FilesystemRepoStore struct {
	mu      sync.Mutex
	m       map[string]*Repository
	datadir string
}

// NewFilesystemRepoStore instantiates a new repo store
func NewFilesystemRepoStore(dir string) *FilesystemRepoStore {
	return &FilesystemRepoStore{
		datadir: dir,
		m:       map[string]*Repository{},
	}
}

// GetRepo with the given id
func (mrs *FilesystemRepoStore) GetRepo(id string) (*Repository, error) {
	if v, ok := mrs.m[id]; ok {
		return v, nil
	}
	return nil, ErrNotFound
}

// CreateRepo with the given repo data
func (mrs *FilesystemRepoStore) CreateRepo(repo *Repository) error {
	mrs.mu.Lock()
	defer mrs.mu.Unlock()

	if _, ok := mrs.m[repo.ID]; ok {
		return ErrExists
	}
	os.MkdirAll(filepath.Join(mrs.datadir, repo.ID), 0755)
	mrs.m[repo.ID] = repo
	return nil
}

// UpdateRepo with the given data
func (mrs *FilesystemRepoStore) UpdateRepo(repo *Repository) error {
	mrs.mu.Lock()
	defer mrs.mu.Unlock()

	if _, ok := mrs.m[repo.ID]; ok {
		mrs.m[repo.ID] = repo
		return nil
	}
	return ErrNotFound
}

// RemoveRepo with the given id
func (mrs *FilesystemRepoStore) RemoveRepo(id string) error {
	mrs.mu.Lock()
	defer mrs.mu.Unlock()

	if _, ok := mrs.m[id]; ok {
		delete(mrs.m, id)
		return nil
	}
	return ErrNotFound
}
