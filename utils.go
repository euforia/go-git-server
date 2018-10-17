package gitserver

import (
	"net/http"
	"strings"
)

func isListRefRequest(r *http.Request) (repo string, service string, ok bool) {
	ss := r.URL.Query()["service"]
	if len(ss) < 1 || (ss[0] != gitRecvPack && ss[0] != gitUploadPack) {
		return
	}
	service = ss[0]

	if strings.HasSuffix(r.URL.Path, "/info/refs") {
		repo = strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/info/refs"), "/")
		ok = true
	}

	return
}

func isPackfileRequest(r *http.Request) (repo string, service string, ok bool) {
	if r.URL.Path == "/"+gitRecvPack || r.URL.Path == "/"+gitUploadPack {
		return
	}

	switch {
	case strings.HasSuffix(r.URL.Path, "/"+gitRecvPack):
		repo = strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"+gitRecvPack), "/")
		service = gitRecvPack
		ok = true

	case strings.HasSuffix(r.URL.Path, "/"+gitUploadPack):
		repo = strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"+gitUploadPack), "/")
		service = gitUploadPack
		ok = true
	}

	return
}

func isUIRequest(r *http.Request) bool {
	agent := r.Header.Get("User-Agent")
	switch {
	case strings.Contains(agent, "Chrome"),
		strings.Contains(agent, "Safari"),
		strings.Contains(agent, "FireFox"),
		strings.Contains(agent, "Mozilla"):
		return true
	}

	return false
}
