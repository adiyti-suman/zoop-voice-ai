package workflow

import (
	"encoding/json"
	"net/http"
	"strings"

	wf "verification-platform/internal/domain/workflow"
	svc "verification-platform/internal/service/workflow"
)

// Handler serves workflow HTTP endpoints.
type Handler struct {
	svc svc.Service
}

func NewHandler(s svc.Service) *Handler {
	return &Handler{svc: s}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func badRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func notFound(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusNotFound, map[string]string{"error": msg})
}

func conflict(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusConflict, map[string]string{"error": msg})
}

func internalError(w http.ResponseWriter) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func domainError(w http.ResponseWriter, err error) {
	switch err {
	case wf.ErrWorkflowNotFound, wf.ErrVersionNotFound, wf.ErrRunNotFound:
		notFound(w, err.Error())
	case wf.ErrInvalidRequest, wf.ErrInvalidGraph:
		badRequest(w, err.Error())
	case wf.ErrVersionNotPublished, wf.ErrInvalidTransition, wf.ErrRunAlreadyTerminal:
		conflict(w, err.Error())
	default:
		internalError(w)
	}
}

// ─── Workflow endpoints ───────────────────────────────────────────────────────

// POST /api/v1/workflows
func (h *Handler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	var req svc.CreateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}
	result, err := h.svc.CreateWorkflow(r.Context(), req)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

// GET /api/v1/workflows/{id}
func (h *Handler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/v1/workflows/")
	result, err := h.svc.GetWorkflow(r.Context(), id)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /api/v1/workflows/{id}/versions
func (h *Handler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	workflowID := extractSegment(r.URL.Path, "/api/v1/workflows/", "/versions")
	var req svc.CreateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}
	result, err := h.svc.CreateVersion(r.Context(), workflowID, req)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

// POST /api/v1/workflow-versions/{id}/validate
func (h *Handler) ValidateVersion(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/v1/workflow-versions/")
	id = strings.TrimSuffix(id, "/validate")
	result, err := h.svc.ValidateVersion(r.Context(), id)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /api/v1/workflow-versions/{id}/publish
func (h *Handler) PublishVersion(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/v1/workflow-versions/")
	id = strings.TrimSuffix(id, "/publish")
	result, err := h.svc.PublishVersion(r.Context(), id)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// POST /api/v1/workflow-versions/{id}/runs
func (h *Handler) StartRun(w http.ResponseWriter, r *http.Request) {
	versionID := extractID(r.URL.Path, "/api/v1/workflow-versions/")
	versionID = strings.TrimSuffix(versionID, "/runs")
	var req svc.StartRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}
	run, err := h.svc.StartRun(r.Context(), versionID, req)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, run)
}

// GET /api/v1/workflow-runs/{id}
func (h *Handler) GetRun(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/v1/workflow-runs/")
	run, nodeRuns, err := h.svc.GetRun(r.Context(), id)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"run":       run,
		"node_runs": nodeRuns,
	})
}

// POST /api/v1/workflow-runs/{id}/cancel
func (h *Handler) CancelRun(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/v1/workflow-runs/")
	id = strings.TrimSuffix(id, "/cancel")
	if err := h.svc.CancelRun(r.Context(), id); err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// ─── URL helpers ──────────────────────────────────────────────────────────────

func extractID(path, prefix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	// Return everything up to the next '/'
	if idx := strings.Index(trimmed, "/"); idx >= 0 {
		return trimmed[:idx]
	}
	return trimmed
}

func extractSegment(path, prefix, suffix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	return strings.TrimSuffix(trimmed, suffix)
}
