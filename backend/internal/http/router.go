package http

import (
	"net/http"
	"strings"

	verfHandler "verification-platform/internal/handler/verification"
	wfHandler "verification-platform/internal/handler/workflow"
)

type Router struct {
	mux        *http.ServeMux
	verfHandler *verfHandler.Handler
	wfHandler  *wfHandler.Handler
}

func NewRouter(vh *verfHandler.Handler, wh *wfHandler.Handler) *Router {
	r := &Router{
		mux:        http.NewServeMux(),
		verfHandler: vh,
		wfHandler:  wh,
	}
	r.registerRoutes()
	return r
}

func (r *Router) registerRoutes() {
	// ── Verification routes ────────────────────────────────────────────────
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

	// ── Workflow routes ────────────────────────────────────────────────────
	r.mux.HandleFunc("/api/v1/workflows", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			r.wfHandler.CreateWorkflow(w, req)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	r.mux.HandleFunc("/api/v1/workflows/", func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/versions") {
			if req.Method == http.MethodPost {
				r.wfHandler.CreateVersion(w, req)
				return
			}
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if req.Method == http.MethodGet {
			r.wfHandler.GetWorkflow(w, req)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// workflow-versions: validate + publish + start run
	r.mux.HandleFunc("/api/v1/workflow-versions/", func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		switch {
		case strings.HasSuffix(path, "/validate") && req.Method == http.MethodPost:
			r.wfHandler.ValidateVersion(w, req)
		case strings.HasSuffix(path, "/publish") && req.Method == http.MethodPost:
			r.wfHandler.PublishVersion(w, req)
		case strings.HasSuffix(path, "/runs") && req.Method == http.MethodPost:
			r.wfHandler.StartRun(w, req)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// workflow-runs: get + cancel
	r.mux.HandleFunc("/api/v1/workflow-runs/", func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path
		switch {
		case strings.HasSuffix(path, "/cancel") && req.Method == http.MethodPost:
			r.wfHandler.CancelRun(w, req)
		case req.Method == http.MethodGet:
			r.wfHandler.GetRun(w, req)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
