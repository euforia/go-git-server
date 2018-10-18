package transport

import (
	"context"
	"log"
	"net/http"

	"github.com/euforia/go-git-server/packproto"
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

// HTTPTransport handles routing git requests leaving the rest alone
type HTTPTransport struct {
	git GitHandler
	// repo handler
	repo http.Handler
	// ui
	ui http.Handler
}

// NewHTTPTransport given the git handler
func NewHTTPTransport(gh GitHandler, rh http.Handler) *HTTPTransport {
	return &HTTPTransport{git: gh, repo: rh}
}

// UIHandler registers the ui handler
func (server *HTTPTransport) UIHandler(h http.Handler) {
	server.ui = h
}

// ServeHTTP assign context to requests and ID to all requests.
func (server *HTTPTransport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("[http] %s %s", r.Method, r.URL.RequestURI())

	switch r.Method {
	case "GET":
		if repoID, service, ok := isListRefRequest(r); ok {
			ctx := context.WithValue(r.Context(), ctxKeyService, service)
			ctx = context.WithValue(ctx, ctxKeyRepo, repoID)
			server.git.ListReferences(w, r.WithContext(ctx))
			return
		}

	case "POST":
		if repoID, service, ok := isPackfileRequest(r); ok {
			ctx := context.WithValue(r.Context(), ctxKeyRepo, repoID)
			switch service {
			case packproto.GitRecvPack:
				server.git.ReceivePack(w, r.WithContext(ctx))
			case packproto.GitUploadPack:
				server.git.UploadPack(w, r.WithContext(ctx))
			default:
				w.WriteHeader(404)
			}
			return
		}

	}

	repoID := r.URL.Path[1:]
	ctx := context.WithValue(r.Context(), ctxKeyRepo, repoID)

	if isUIRequest(r) && server.ui != nil {
		server.ui.ServeHTTP(w, r.WithContext(ctx))
		return
	}

	server.repo.ServeHTTP(w, r.WithContext(ctx))
}

// ListenAndServe starts and listener on addr and serves the router handlers
func (server *HTTPTransport) ListenAndServe(addr string) error {
	log.Printf("HTTP Server: http://%s", addr)
	return http.ListenAndServe(addr, server)
}
