package repository

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

// RepositoryHead contains the HEAD ref and hash information
type RepositoryHead struct {
	Ref  string        `json:"ref"`
	Hash plumbing.Hash `json:"hash"`
}

// RepositoryReferences contains repo refs and head information.
type RepositoryReferences struct {
	mu    sync.Mutex
	head  RepositoryHead
	Heads map[string]plumbing.Hash
	Tags  map[string]plumbing.Hash
}

// NewRepositoryReferences instantiates a new RepositoryReferences structure with
// the defaults.
func NewRepositoryReferences() *RepositoryReferences {
	return &RepositoryReferences{
		head:  RepositoryHead{Ref: "heads/master", Hash: plumbing.Hash{}},
		Heads: map[string]plumbing.Hash{"master": plumbing.Hash{}},
		Tags:  map[string]plumbing.Hash{},
	}
}

// Head returns the current pointed HEAD
func (refs *RepositoryReferences) Head() RepositoryHead {
	return refs.head
}

// UpdateRef updates a repo reference given the previous hash of the ref.
func (refs *RepositoryReferences) UpdateRef(ref string, prev, curr plumbing.Hash) error {
	s := strings.Split(ref, "/")
	s = s[1:]

	refs.mu.Lock()
	defer refs.mu.Unlock()

	switch s[0] {
	case "heads":

		v, ok := refs.Heads[s[1]]
		if !ok {
			return fmt.Errorf("ref not found: %s", ref)
		}

		if v.String() == prev.String() {
			refs.Heads[s[1]] = curr
			break
		}

		return fmt.Errorf("previous hash mismatch: %s != %s", v.String(), prev.String())

	case "tags":

		v, ok := refs.Tags[s[1]]
		if !ok {
			return fmt.Errorf("ref not found: %s", ref)
		}

		if v.String() == prev.String() {
			refs.Tags[s[1]] = curr
			break
		}

		return fmt.Errorf("previous hash mismatch: %s != %s", v.String(), prev.String())

	default:
		return fmt.Errorf("invalid ref: %s", ref)

	}

	return nil
}

// MarshalJSON is a custom json marshaller for the repository specifically to handle
// hashes.
func (refs *RepositoryReferences) MarshalJSON() ([]byte, error) {
	out := map[string]interface{}{
		"head": map[string]string{
			"ref":  refs.head.Ref,
			"hash": refs.head.Hash.String(),
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
			refs.head = RepositoryHead{Hash: h, Ref: ref}
			return h, nil
		}

	case "heads":
		if h, ok := refs.Heads[tr[1]]; ok {
			refs.head = RepositoryHead{Hash: h, Ref: ref}
			return h, nil
		}

	}

	return plumbing.Hash{}, fmt.Errorf("invalid ref: %s", ref)
}
