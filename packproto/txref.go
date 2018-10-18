package packproto

import (
	"errors"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

// TxRef is a transaction to update a repo reference
type txRef struct {
	oldHash plumbing.Hash
	newHash plumbing.Hash
	ref     string
}

// Parses old hash, new hash, and ref from a line in that order
func newTxRefFromBytes(line []byte) (rt txRef, err error) {
	arr := strings.Split(string(line), " ")
	if len(arr) < 3 {
		err = errors.New("invalid line: " + string(line))
		return
	}

	rt = txRef{
		oldHash: plumbing.NewHash(arr[0]),
		newHash: plumbing.NewHash(arr[1]),
		ref:     strings.TrimSuffix(arr[2], string([]byte{0})),
	}

	return
}

// Old returns the old Reference object
func (tx *txRef) old() *plumbing.Reference {
	return plumbing.NewHashReference(plumbing.ReferenceName(tx.ref), tx.oldHash)
}

// New returns the new Reference object
func (tx *txRef) new() *plumbing.Reference {
	return plumbing.NewHashReference(plumbing.ReferenceName(tx.ref), tx.newHash)
}
