package gitserver

import (
	"context"
	"log"
	"net/http"
	"strings"
)

const (
	serviceRecvPack   = "git-receive-pack"
	serviceUploadPack = "git-upload-pack"
)

// GitHandler interface for git specific operations
type GitHandler interface {
	// clone, fetch, pull ???
	UploadPack(w http.ResponseWriter, r *http.Request)
	// push
	ReceivePack(w http.ResponseWriter, r *http.Request)
	// list refs based on receive or upload pack
	ListReferences(w http.ResponseWriter, r *http.Request)
}

// Router handles routing git requests leaving the rest alone
type Router struct {
	GitHandler
	// Default http handler when not a git request or handled one
	Default func(w http.ResponseWriter, r *http.Request)
}

// NewRouter given the git handler
func NewRouter(gh GitHandler) *Router {
	return &Router{GitHandler: gh}
}

// ServeHTTP assign context to requests and ID to all requests.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[http] %s %s", r.Method, r.URL.RequestURI())
	// TODO: Add auth.

	switch r.Method {
	case "GET":
		if repoID, service, ok := isListRefRequest(r); ok {
			ctx := context.WithValue(r.Context(), "service", service)
			ctx = context.WithValue(ctx, "ID", repoID)
			router.ListReferences(w, r.WithContext(ctx))
			return
		}

	case "POST":
		if repoID, service, ok := isPackfileRequest(r); ok {
			ctx := context.WithValue(r.Context(), "ID", repoID)
			if service == serviceRecvPack {
				router.ReceivePack(w, r.WithContext(ctx))
			} else {
				router.UploadPack(w, r.WithContext(ctx))
			}
			return
		}

	}
	ctx := context.WithValue(r.Context(), "ID", strings.TrimPrefix(r.URL.Path, "/"))
	router.Default(w, r.WithContext(ctx))
}

func isListRefRequest(r *http.Request) (repo string, service string, ok bool) {
	ss, ok := r.URL.Query()["service"]
	if !ok || len(ss) < 1 || (ss[0] != serviceRecvPack && ss[0] != serviceUploadPack) {
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

func isPackfileRequest(r *http.Request) (repo string, service string, ok bool) {
	if r.URL.Path == "/"+serviceRecvPack || r.URL.Path == "/"+serviceUploadPack {
		return
	}

	switch {
	case strings.HasSuffix(r.URL.Path, "/"+serviceRecvPack):
		repo = strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"+serviceRecvPack), "/")
		service = serviceRecvPack
		ok = true

	case strings.HasSuffix(r.URL.Path, "/"+serviceUploadPack):
		repo = strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"+serviceUploadPack), "/")
		service = serviceUploadPack
		ok = true
	}

	return
}
