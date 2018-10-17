package main

import (
	"flag"
	"log"
	"os"

	"github.com/euforia/go-git-server"
	"github.com/euforia/go-git-server/repository"
)

var (
	httpAddr = "127.0.0.1:12345"
	dataDir  = flag.String("data-dir", "./tmp", "dir")
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	flag.Parse()

	repostore := repository.NewFilesystemRepoStore(*dataDir)
	// repostore := repository.NewMemRepoStore()

	os.MkdirAll(*dataDir, 0755)
	objstores := gitserver.NewFilesystemObjectStorage(*dataDir)
	// objstores := gitserver.NewMemObjectStorage()

	gh := gitserver.NewGitHTTPService(objstores)
	rh := gitserver.NewRepoHTTPService(repostore)

	router := gitserver.NewRouter(gh, rh, nil)
	if err := router.Serve(httpAddr); err != nil {
		log.Println(err)
	}
}
