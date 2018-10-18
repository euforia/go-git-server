package repository

import (
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"
)

// Manager wraps an underlying repo store with a git repo manager
// to initialize repos
type Manager struct {
	RepositoryStore
	rmgr *GitRepoManager
}

func NewManager(s RepositoryStore, mgr *GitRepoManager) *Manager {
	return &Manager{RepositoryStore: s, rmgr: mgr}
}

func (s *Manager) CreateRepo(repo *Repository) error {
	err := s.rmgr.CreateRepo(repo.ID)
	if err == nil {
		err = s.RepositoryStore.CreateRepo(repo)
	}
	return err
}

type GitRepoManager struct {
	datadir string
}

func NewGitRepoManager(dir string) *GitRepoManager {
	return &GitRepoManager{datadir: dir}
}

func (m *GitRepoManager) CreateRepo(id string) error {
	path := filepath.Join(m.datadir, id)
	_, err := os.Stat(path)
	if err == nil {
		return ErrExists
	}
	_, err = git.PlainInit(path, true)
	return err
}

func (m *GitRepoManager) GetRepo(id string) (*git.Repository, error) {
	path := filepath.Join(m.datadir, id)
	return git.PlainOpen(path)
}

func (m *GitRepoManager) RemoveRepo(id string) error {
	path := filepath.Join(m.datadir, id)
	_, err := os.Stat(path)
	if err != nil {
		return ErrNotFound
	}

	return os.RemoveAll(path)
}
