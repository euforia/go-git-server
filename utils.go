package gitserver

import (
	"fmt"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Parses old hash, new hash, and ref from a line
func parseReceivePackHeader(line []byte) (oldH, newH plumbing.Hash, ref string, err error) {
	s := string(line)
	arr := strings.Split(s, " ")
	if len(arr) < 3 {
		err = fmt.Errorf("invalid line: %s", line)
		return
	}

	oldH = plumbing.NewHash(arr[0])
	newH = plumbing.NewHash(arr[1])
	ref = strings.TrimSuffix(arr[2], string([]byte{0}))
	return
}

func updateRepoRefsFromLine(line []byte, repo *Repository) (string, error) {
	_, nh, ref, err := parseReceivePackHeader(line)
	if err != nil {
		return "", err
	}

	//
	// TODO:
	// - Create transaction to update ref.
	// log.Printf("[tx] old=%s new=%s ref=%s", oh, nh, ref)

	s := strings.Split(ref, "/")
	s = s[1:]

	switch s[0] {
	case "heads":
		// TODO: check oh
		repo.Refs.Heads[s[1]] = nh
	case "tags":
		// TODO: check oh
		repo.Refs.Tags[s[1]] = nh

	}

	// Set head if ref. matches
	pref := strings.Join(s, "/")
	if repo.Refs.Head.Ref == pref {
		if _, err = repo.Refs.SetHead(pref); err != nil {
			return "", err
		}
	}

	return ref, nil
}
