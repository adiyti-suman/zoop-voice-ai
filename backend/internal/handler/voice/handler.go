package voice

import (
	"encoding/json"
	"net/http"
	"strings"

	domain "verification-platform/internal/domain/voice"
	svc "verification-platform/internal/service/voice"
)

// Handler serves Voice session REST endpoints.
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
	case domain.ErrSessionNotFound:
		notFound(w, err.Error())
	case domain.ErrInvalidAudio:
		badRequest(w, err.Error())
	case domain.ErrInvalidTransition:
		conflict(w, err.Error())
	default:
		internalError(w)
	}
}

// ─── Endpoints ────────────────────────────────────────────────────────────────

// POST /api/v1/voice/sessions
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req svc.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "invalid request body")
		return
	}
	session, err := h.svc.CreateSession(r.Context(), req)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

// GET /api/v1/voice/sessions/{id}
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/v1/voice/sessions/")
	session, err := h.svc.GetSession(r.Context(), id)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, session)
}

// POST /api/v1/voice/sessions/{id}/pause
func (h *Handler) PauseSession(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/v1/voice/sessions/")
	id = strings.TrimSuffix(id, "/pause")
	if err := h.svc.TransitionStatus(r.Context(), id, domain.SessionStatusPaused); err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "PAUSED"})
}

// POST /api/v1/voice/sessions/{id}/resume
func (h *Handler) ResumeSession(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/v1/voice/sessions/")
	id = strings.TrimSuffix(id, "/resume")
	if err := h.svc.TransitionStatus(r.Context(), id, domain.SessionStatusListening); err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "LISTENING"})
}

// POST /api/v1/voice/sessions/{id}/stop
func (h *Handler) StopSession(w http.ResponseWriter, r *http.Request) {
	id := extractID(r.URL.Path, "/api/v1/voice/sessions/")
	id = strings.TrimSuffix(id, "/stop")
	// For REST, stopping means COMPLETED generally unless failed
	if err := h.svc.TransitionStatus(r.Context(), id, domain.SessionStatusCompleted); err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "COMPLETED"})
}

func extractID(path, prefix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	if idx := strings.Index(trimmed, "/"); idx >= 0 {
		return trimmed[:idx]
	}
	return trimmed
}
