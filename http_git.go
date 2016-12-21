package gitserver

import (
	"fmt"
	"net/http"

	"github.com/euforia/go-git-server/repository"

	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

// GitHTTPService is a git http server
type GitHTTPService struct {
	// object store
	store storer.EncodedObjectStorer
	// repository store
	repos repository.RepositoryStore
}

// NewGitHTTPService instantiates the git http service with the provided repo store
// and object store.
func NewGitHTTPService(repostore repository.RepositoryStore, objstore storer.EncodedObjectStorer) *GitHTTPService {
	svr := &GitHTTPService{
		store: objstore,
		repos: repostore,
	}

	return svr
}

// ListReferences per the git protocol
func (svr *GitHTTPService) ListReferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repoID := ctx.Value("ID").(string)
	service := ctx.Value("service").(string)

	repo, err := svr.repos.GetRepo(repoID)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
	w.WriteHeader(200)

	proto := NewProtocol(w, nil)
	proto.ListReferences(GitServiceType(service), repo.Refs)
}

// ReceivePack implements the receive-pack protocol over http
func (svr *GitHTTPService) ReceivePack(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repoID := ctx.Value("ID").(string)

	repo, err := svr.repos.GetRepo(repoID)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	defer r.Body.Close()

	w.Header().Add("Content-Type", "application/x-git-receive-pack-result")
	w.WriteHeader(200)

	proto := NewProtocol(w, r.Body)
	proto.ReceivePack(repo, svr.repos, svr.store)
	// TODO: Call hooks
}

// UploadPack implements upload-pack protocol over http
func (svr *GitHTTPService) UploadPack(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repoID := ctx.Value("ID").(string)

	if _, err := svr.repos.GetRepo(repoID); err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	defer r.Body.Close()

	proto := NewProtocol(w, r.Body)
	proto.UploadPack(svr.store)
}
