package gitserver

import (
	"net/http"
	"strings"
)

func isListRefRequest(r *http.Request) (repo string, service string, ok bool) {
	ss, ok := r.URL.Query()["service"]
	if !ok || len(ss) < 1 || (GitServiceType(ss[0]) != GitServiceRecvPack && GitServiceType(ss[0]) != GitServiceUploadPack) {
		return
	}
	service = ss[0]

	// not list ref repo info not there
	if r.URL.Path == "/info/refs" || !strings.HasSuffix(r.URL.Path, "info/refs") {
		return
	}

	repo = strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/info/refs"), "/")
	ok = true
	return
}

func isPackfileRequest(r *http.Request) (repo string, service GitServiceType, ok bool) {
	if r.URL.Path == "/"+string(GitServiceRecvPack) || r.URL.Path == "/"+string(GitServiceUploadPack) {
		return
	}

	switch {
	case strings.HasSuffix(r.URL.Path, "/"+string(GitServiceRecvPack)):
		repo = strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"+string(GitServiceRecvPack)), "/")
		service = GitServiceRecvPack
		ok = true

	case strings.HasSuffix(r.URL.Path, "/"+string(GitServiceUploadPack)):
		repo = strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"+string(GitServiceUploadPack)), "/")
		service = GitServiceUploadPack
		ok = true
	}

	return
}

func isUIRequest(r *http.Request) bool {
	agent := r.Header.Get("User-Agent")
	//log.Printf("%+v", usrAgt)
	switch {
	case strings.Contains(agent, "Chrome"),
		strings.Contains(agent, "Safari"),
		strings.Contains(agent, "FireFox"):
		return true
	}

	return false
}
