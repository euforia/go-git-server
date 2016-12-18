package main

import (
	"log"
	"net/http"

	"github.com/euforia/go-git-server"
)

var (
	httpAddr = "127.0.0.1:12345"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	gh := gitserver.NewGitServer()
	router := gitserver.NewRouter(gh)
	router.Default = gh.HandleRepository

	log.Printf("HTTP Server: http://%s", httpAddr)
	if err := http.ListenAndServe(httpAddr, router); err != nil {
		log.Println(err)
	}
}
