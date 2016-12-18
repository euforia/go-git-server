package main

import (
	"log"

	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/euforia/go-git-server"
	"github.com/euforia/go-git-server/repository"
)

var (
	httpAddr = "127.0.0.1:12345"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	repostore := repository.NewMemRepoStore()
	objstore := memory.NewStorage()

	gh := gitserver.NewGitHTTPService(repostore, objstore)
	rh := gitserver.NewRepoHTTPService(repostore)

	router := gitserver.NewRouter(gh, rh, nil)
	if err := router.Serve(httpAddr); err != nil {
		log.Println(err)
	}
}
