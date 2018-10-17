package gitserver

import (
	"fmt"
	"net/http"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

// GitHTTPService is a git http server
type GitHTTPService struct {
	// Store containing all repo storage
	stores ObjectStorage
}

// NewGitHTTPService instantiates the git http service with the provided repo store
// and object store.
func NewGitHTTPService(objstore ObjectStorage) *GitHTTPService {
	svr := &GitHTTPService{
		stores: objstore,
	}

	return svr
}

// ListReferences per the git protocol
func (svr *GitHTTPService) ListReferences(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ctx := r.Context()
	repoID := ctx.Value(ctxKeyRepo).(string)
	service := ctx.Value(ctxKeyService).(string)

	st := svr.stores.GetStore(repoID)
	if st == nil {
		w.WriteHeader(404)
		return
	}
	riter, err := st.IterReferences()
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	refs := make([]*plumbing.Reference, 0)
	riter.ForEach(func(ref *plumbing.Reference) error {
		refs = append(refs, ref)
		return nil
	})

	w.Header().Add("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
	w.WriteHeader(200)

	proto := NewProtocol(w, nil)
	proto.ListReferences(service, refs)
}

// ReceivePack implements the receive-pack protocol over http
func (svr *GitHTTPService) ReceivePack(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	repoID := r.Context().Value(ctxKeyRepo).(string)
	st := svr.stores.GetStore(repoID)
	if st == nil {
		w.WriteHeader(404)
		return
	}

	w.Header().Add("Content-Type", "application/x-git-receive-pack-result")
	w.WriteHeader(200)

	proto := NewProtocol(w, r.Body)
	proto.ReceivePack(st)
}

// UploadPack implements upload-pack protocol over http
func (svr *GitHTTPService) UploadPack(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	repoID := r.Context().Value(ctxKeyRepo).(string)
	st := svr.stores.GetStore(repoID)
	if st == nil {
		w.WriteHeader(404)
		return
	}

	proto := NewProtocol(w, r.Body)
	proto.UploadPack(st)
}
