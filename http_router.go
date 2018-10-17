package gitserver

import (
	"context"
	"log"
	"net/http"
)

type ctxKey string

const (
	ctxKeyService ctxKey = "service"
	ctxKeyRepo    ctxKey = "repo"
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
	git GitHandler
	// repo handler
	repo http.Handler
	// ui
	ui http.Handler
}

// NewRouter given the git handler
func NewRouter(gh GitHandler, rh http.Handler, uh http.Handler) *Router {
	return &Router{git: gh, repo: rh, ui: uh}
}

// ServeHTTP assign context to requests and ID to all requests.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[http] %s %s", r.Method, r.URL.RequestURI())

	switch r.Method {
	case "GET":
		if repoID, service, ok := isListRefRequest(r); ok {
			ctx := context.WithValue(r.Context(), ctxKeyService, service)
			ctx = context.WithValue(ctx, ctxKeyRepo, repoID)
			router.git.ListReferences(w, r.WithContext(ctx))
			return
		}

	case "POST":
		if repoID, service, ok := isPackfileRequest(r); ok {
			ctx := context.WithValue(r.Context(), ctxKeyRepo, repoID)
			switch service {
			case gitRecvPack:
				router.git.ReceivePack(w, r.WithContext(ctx))
			case gitUploadPack:
				router.git.UploadPack(w, r.WithContext(ctx))
			default:
				w.WriteHeader(404)
			}
			return
		}

	}

	repoID := r.URL.Path[1:]
	ctx := context.WithValue(r.Context(), ctxKeyRepo, repoID)

	if isUIRequest(r) && router.ui != nil {
		router.ui.ServeHTTP(w, r.WithContext(ctx))
		return
	}

	router.repo.ServeHTTP(w, r.WithContext(ctx))
}

// Serve starts serving the router handlers
func (router *Router) Serve(addr string) error {
	log.Printf("[git] HTTP Server: http://%s", addr)
	return http.ListenAndServe(addr, router)
}
