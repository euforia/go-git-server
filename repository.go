package gitserver

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

//  RepositoryHead contains the HEAD ref and hash information
type RepositoryHead struct {
	Ref  string        `json:"ref"`
	Hash plumbing.Hash `json:"hash"`
}

// RepositoryReferences contains repo refs and head information.
type RepositoryReferences struct {
	mu    sync.Mutex
	Head  RepositoryHead
	Heads map[string]plumbing.Hash
	Tags  map[string]plumbing.Hash
}

// NewRepositoryReferences instantiates a new RepositoryReferences structure with
// the defaults.
func NewRepositoryReferences() *RepositoryReferences {
	return &RepositoryReferences{
		Head:  RepositoryHead{Ref: "heads/master", Hash: plumbing.Hash{}},
		Heads: map[string]plumbing.Hash{"master": plumbing.Hash{}},
		Tags:  map[string]plumbing.Hash{},
	}
}

// MarshalJSON is a custom json marshaller for the repository specifically to handle
// hashes.
func (refs *RepositoryReferences) MarshalJSON() ([]byte, error) {
	out := map[string]interface{}{
		"head": map[string]string{
			"ref":  refs.Head.Ref,
			"hash": refs.Head.Hash.String(),
		},
	}
	heads := map[string]string{}
	for k, v := range refs.Heads {
		heads[k] = v.String()
	}
	out["heads"] = heads

	tags := map[string]string{}
	for k, v := range refs.Tags {
		tags[k] = v.String()
	}
	out["tags"] = tags

	return json.Marshal(out)
}

// SetHead given the ref.  Returns an error if the ref is not found or invalid.
func (refs *RepositoryReferences) SetHead(ref string) (plumbing.Hash, error) {
	tr := strings.Split(ref, "/")
	if len(tr) != 2 {
		return plumbing.Hash{}, fmt.Errorf("invalid ref: %s", ref)
	}

	refs.mu.Lock()
	defer refs.mu.Unlock()

	switch tr[0] {
	case "tags":
		if h, ok := refs.Tags[tr[1]]; ok {
			refs.Head = RepositoryHead{Hash: h, Ref: ref}
			return h, nil
		}

	case "heads":
		if h, ok := refs.Heads[tr[1]]; ok {
			refs.Head = RepositoryHead{Hash: h, Ref: ref}
			return h, nil
		}

	}

	return plumbing.Hash{}, fmt.Errorf("invalid ref: %s", ref)
}

// Repository represents a single repo.
type Repository struct {
	ID   string                `json:"id"`
	Refs *RepositoryReferences `json:"refs"`
	// sha256 hash of the binary data from the store.  This is stored upon
	// retreival and used as the previous hash when trying to update.
	hash [32]byte
}

// NewRepository instantiates an empty repo.
func NewRepository(id string) *Repository {
	return &Repository{ID: id, Refs: NewRepositoryReferences()}
}

// Hash of the repo as seen by the datastore.  This is used as the previous hash
// to issue any updates back to the datastore.
func (repo *Repository) Hash() [32]byte {
	return repo.hash
}

func (repo *Repository) String() string {
	return repo.ID
}
