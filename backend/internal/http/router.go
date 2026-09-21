package http

import (
	"net/http"
	"strings"
	"verification-platform/internal/handler/verification"
)

type Router struct {
	mux         *http.ServeMux
	verfHandler *verification.Handler
}

func NewRouter(verfHandler *verification.Handler) *Router {
	r := &Router{
		mux:         http.NewServeMux(),
		verfHandler: verfHandler,
	}
	r.registerRoutes()
	return r
}

func (r *Router) registerRoutes() {
	r.mux.HandleFunc("/api/v1/verifications", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			r.verfHandler.CreateVerification(w, req)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	r.mux.HandleFunc("/api/v1/verifications/", func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/transition") {
			if req.Method == http.MethodPost {
				r.verfHandler.TransitionVerification(w, req)
				return
			}
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if req.Method == http.MethodGet {
			r.verfHandler.GetVerification(w, req)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
