package gitserver

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bargez/pktline"
	"github.com/euforia/go-git-server/packfile"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// RepositoryStore implements a repository storage
type RepositoryStore interface {
	CreateRepo(*Repository) error
	UpdateRepo(*Repository) error
	GetRepo(string) (*Repository, error)
	//RemoveRepo(string) error
}

type GitServer struct {
	store *memory.Storage
	// repository store
	repos RepositoryStore
}

func NewGitServer() *GitServer {
	svr := &GitServer{
		store: memory.NewStorage(),
		repos: NewMemRepoStore(),
	}

	return svr
}

// ListReferences per the git protocol
func (svr *GitServer) ListReferences(w http.ResponseWriter, r *http.Request) {
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

	// Start sending info
	enc := pktline.NewEncoder(w)
	enc.Encode([]byte(fmt.Sprintf("# service=%s\n", service)))
	enc.Encode(nil)

	// Repo empty so send zeros
	if (repo.Refs.Heads == nil || len(repo.Refs.Heads) == 0) && (repo.Refs.Tags == nil || len(repo.Refs.Tags) == 0) {
		b0 := append([]byte("0000000000000000000000000000000000000000"), 32)
		b0 = append(b0, nullCapabilities()...)

		enc.Encode(append(b0, 10))
		enc.Encode(nil)
		return
	}

	// Send HEAD info
	head := repo.Refs.Head

	lh := append([]byte(fmt.Sprintf("%s HEAD", head.Hash.String())), '\x00')
	lh = append(lh, capabilities()...)

	if service == serviceUploadPack {
		lh = append(lh, []byte(" symref=HEAD:refs/"+head.Ref)...)
	}
	enc.Encode(append(lh, 10))

	// Send refs - heads
	for href, h := range repo.Refs.Heads {
		enc.Encode([]byte(fmt.Sprintf("%s refs/heads/%s\n", h.String(), href)))
	}

	// Send refs - tags
	for tref, h := range repo.Refs.Tags {
		enc.Encode([]byte(fmt.Sprintf("%s refs/tags/%s\n", h.String(), tref)))
	}

	enc.Encode(nil)
}

func (svr *GitServer) ReceivePack(w http.ResponseWriter, r *http.Request) {
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

	var (
		lines [][]byte
		dec   = pktline.NewDecoder(r.Body)
	)
	// Read refs from client
	if e := dec.DecodeUntilFlush(&lines); e != nil {
		log.Printf("[receive-pack] ERR %v", e)
	}

	for _, l := range lines {
		log.Printf("[receive-pack] DBG %s", l)
	}

	enc := pktline.NewEncoder(w)

	packdec := packfile.NewDecoder(r.Body, svr.store)
	if err = packdec.Decode(); err != nil {
		enc.Encode([]byte(fmt.Sprintf("unpack %v", err)))
		return
	}
	enc.Encode([]byte("unpack ok\n"))

	// Send status for each ref.
	for _, l := range lines {
		ref, err := updateRepoRefsFromLine(l, repo)
		if err == nil {
			enc.Encode([]byte(fmt.Sprintf("ok %s\n", ref)))
			continue
		}
		log.Println("ERR", err)
	}

	if err := svr.repos.UpdateRepo(repo); err != nil {
		log.Println("[receive-pack] ERR Failed to update repo:", err)
	}

	enc.Encode(nil)
}

func (svr *GitServer) UploadPack(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repoID := ctx.Value("ID").(string)

	_, err := svr.repos.GetRepo(repoID)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
		return
	}

	defer r.Body.Close()

	var (
		wants []string
		haves []string
		dec   = pktline.NewDecoder(r.Body)
	)

	for {
		var line []byte
		if err = dec.Decode(&line); err != nil {
			break
		} else if len(line) == 0 {
			continue
		} else {
			line = line[:len(line)-1]
		}

		if string(line) == "done" {
			break
		}

		log.Printf("[upload-pack] %s", line)

		op := strings.Split(string(line), " ")
		switch op[0] {
		case "want":
			wants = append(wants, op[1])

		case "have":
			haves = append(haves, op[1])

		}
	}

	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	log.Printf("[upload-pack] wants=%d haves=%d", len(wants), len(haves))

	// object map to hold uniques
	out := map[plumbing.Hash]plumbing.EncodedObject{}
	for _, want := range wants {
		o, err1 := svr.store.EncodedObject(plumbing.AnyObject, plumbing.NewHash(want))
		if err1 != nil {
			log.Println("ERR", want, err1)
			continue
		}

		// Walk the wants
		err1 = svr.WalkObject(o, func(obj plumbing.EncodedObject) error {
			//log.Printf("> %s %s %d", obj.Type(), obj.Hash(), obj.Size())
			out[obj.Hash()] = obj
			return nil
		})

		if err1 != nil {
			log.Println("ERR", o.Hash().String(), err1)
		}
	}

	enc := pktline.NewEncoder(w)
	enc.Encode([]byte("NAK\n"))

	// Write pack file
	var chksum []byte
	if chksum, err = packfile.WritePackFile(out, w); err != nil {
		log.Println(err)
	} else {
		log.Printf("[upload-pack] Packfile encoded: checksum=%x, objects=%d", chksum, len(out))
	}
}

// WalkObject walks an object
func (svr *GitServer) WalkObject(obj plumbing.EncodedObject, cb func(plumbing.EncodedObject) error) error {
	err := cb(obj)
	if err != nil {
		return err
	}

	switch obj.Type() {
	case plumbing.CommitObject:
		var commit *object.Commit
		if commit, err = object.GetCommit(svr.store, obj.Hash()); err != nil {
			//log.Println("ERR", obj.Hash().String(), err)
			break
		}

		var tobj *object.Tree
		if tobj, err = commit.Tree(); err != nil {
			break
		}

		var to plumbing.EncodedObject
		if to, err = svr.store.EncodedObject(plumbing.AnyObject, tobj.Hash); err != nil {
			//log.Println("ERR", err)
			break
		}

		if err = svr.WalkObject(to, cb); err != nil {
			break
		}

		iter := commit.Parents()
		for {
			cmt, e1 := iter.Next()
			if e1 != nil {
				if e1 != io.EOF {
					//log.Println("ERR parent", e1)
					err = e1
				}
				break
			}

			var o plumbing.EncodedObject
			if o, e1 = svr.store.EncodedObject(plumbing.AnyObject, cmt.Hash); e1 != nil {
				err = e1
				break
			}

			if e1 = svr.WalkObject(o, cb); e1 != nil {
				//log.Println("ERR", e1)
				err = e1
				break
			}
		}

	case plumbing.TreeObject:
		t := &object.Tree{}
		if err = t.Decode(obj); err != nil {
			break
		}

		for _, entry := range t.Entries {
			po, err1 := svr.store.EncodedObject(plumbing.AnyObject, entry.Hash)
			if err1 == nil {
				if err1 = svr.WalkObject(po, cb); err != nil {
					log.Println("ERR", err1)
				}
			}
		}

	}

	return err
}

func (svr *GitServer) HandleRepository(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	repoID := ctx.Value("ID").(string)

	if !strings.Contains(repoID, "/") {
		w.WriteHeader(404)
		return
	}

	var (
		err  error
		repo *Repository
	)

	switch r.Method {
	case "GET":
		repo, err = svr.repos.GetRepo(repoID)

	case "PUT":
		// Create
		dec := json.NewDecoder(r.Body)
		defer r.Body.Close()

		repo = NewRepository(repoID)
		if err = dec.Decode(&repo); err == nil || err == io.EOF {
			err = svr.repos.CreateRepo(repo)
		}

	case "POST":
		// Update
		dec := json.NewDecoder(r.Body)
		defer r.Body.Close()

		// Get existing
		if repo, err = svr.repos.GetRepo(repoID); err == nil {
			// Unmarshal on to existing
			if err = dec.Decode(repo); err == nil {
				err = svr.repos.UpdateRepo(repo)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
	} else {
		b, _ := json.Marshal(repo)
		w.WriteHeader(200)
		w.Write(b)
	}

}

func capabilities() []byte {
	//return []byte("report-status delete-refs ofs-delta multi_ack_detailed")
	return []byte("report-status delete-refs ofs-delta")
}

func nullCapabilities() []byte {
	return append(append([]byte("capabilities^{}"), '\x00'), capabilities()...)
}
