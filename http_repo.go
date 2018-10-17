package gitserver

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/euforia/go-git-server/repository"
)

type RepoHTTPService struct {
	repos repository.RepositoryStore
}

func NewRepoHTTPService(store repository.RepositoryStore) *RepoHTTPService {
	return &RepoHTTPService{repos: store}
}

func (svr *RepoHTTPService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	repoID := r.Context().Value(ctxKeyRepo).(string)
	if !strings.Contains(repoID, "/") {
		w.WriteHeader(404)
		return
	}

	var (
		code = 400
		err  error
		repo *repository.Repository
	)

	switch r.Method {
	case "GET":
		repo, err = svr.repos.GetRepo(repoID)

	case "PUT":
		// Create
		dec := json.NewDecoder(r.Body)
		defer r.Body.Close()

		repo = repository.NewRepository(repoID)
		if err = dec.Decode(&repo); err == nil || err == io.EOF {
			if err = svr.repos.CreateRepo(repo); err == repository.ErrExists {
				code = 409
			}
		}

	case "POST":
		// Update
		dec := json.NewDecoder(r.Body)
		defer r.Body.Close()

		// Get existing
		if repo, err = svr.repos.GetRepo(repoID); err == nil {
			// Unmarshal on to existing
			if err = dec.Decode(repo); err == nil {
				if err = svr.repos.UpdateRepo(repo); err == repository.ErrNotFound {
					code = 404
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(code)
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
	} else {
		b, _ := json.Marshal(repo)
		w.WriteHeader(200)
		w.Write(b)
	}

}
