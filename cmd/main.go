package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/euforia/go-git-server/repository"
	"github.com/euforia/go-git-server/storage"
	"github.com/euforia/go-git-server/transport"
)

var (
	httpAddr = "127.0.0.1:12345"
	dataDir  = flag.String("data-dir", "", "dir")
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func makeManager() *repository.Manager {
	os.MkdirAll(*dataDir, 0755)
	repoStore := repository.NewFilesystemRepoStore(*dataDir)
	gitRepoMgr := repository.NewGitRepoManager(*dataDir)
	return repository.NewManager(repoStore, gitRepoMgr)
}

func main() {
	flag.Parse()
	if *dataDir == "" {
		fmt.Println("-data-dir required!")
		os.Exit(1)
	}

	objStore := storage.NewFilesystemGitRepoStorage(*dataDir)
	gh := transport.NewGitHTTPService(objStore)

	mgr := makeManager()
	rh := transport.NewRepoHTTPService(mgr)

	server := transport.NewHTTPTransport(gh, rh)
	if err := server.ListenAndServe(httpAddr); err != nil {
		log.Println(err)
	}
}
