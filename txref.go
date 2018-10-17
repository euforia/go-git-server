package gitserver

import (
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// TxRef is a transaction to update a repo reference
type TxRef struct {
	oldHash plumbing.Hash
	newHash plumbing.Hash
	ref     string
}

// Old returns the old Reference object
func (tx *TxRef) Old() *plumbing.Reference {
	return plumbing.NewHashReference(plumbing.ReferenceName(tx.ref), tx.oldHash)
}

// New returns the new Reference object
func (tx *TxRef) New() *plumbing.Reference {
	return plumbing.NewHashReference(plumbing.ReferenceName(tx.ref), tx.newHash)
}
