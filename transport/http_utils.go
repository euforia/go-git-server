package transport

import (
	"net/http"
	"strings"

	"github.com/euforia/go-git-server/packproto"
)

func isListRefRequest(r *http.Request) (repo string, service string, ok bool) {
	ss := r.URL.Query()["service"]
	if len(ss) < 1 || (ss[0] != packproto.GitRecvPack && ss[0] != packproto.GitUploadPack) {
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
	switch {
	case strings.HasSuffix(r.URL.Path, "/"+packproto.GitRecvPack):
		repo = strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"+packproto.GitRecvPack), "/")
		service = packproto.GitRecvPack
		ok = true

	case strings.HasSuffix(r.URL.Path, "/"+packproto.GitUploadPack):
		repo = strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"+packproto.GitUploadPack), "/")
		service = packproto.GitUploadPack
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
