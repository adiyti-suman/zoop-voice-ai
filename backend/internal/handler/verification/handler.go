package verification

import (
	"encoding/json"
	"net/http"
	"strings"

	domain "verification-platform/internal/domain/verification"
	service "verification-platform/internal/service/verification"
)

type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) CreateVerification(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := r.Header.Get("Idempotency-Key")

	var req service.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, domain.ErrInvalidRequest)
		return
	}

	v, err := h.svc.Create(r.Context(), idempotencyKey, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrInvalidVerificationType {
			status = http.StatusBadRequest
		}
		writeError(w, status, err)
		return
	}

	writeJSON(w, http.StatusCreated, v)
}

func (h *Handler) GetVerification(w http.ResponseWriter, r *http.Request) {
	// Minimal router extraction - assuming /api/v1/verifications/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		writeError(w, http.StatusBadRequest, domain.ErrInvalidRequest)
		return
	}
	id := pathParts[4]

	v, err := h.svc.Get(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrVerificationNotFound {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}

	writeJSON(w, http.StatusOK, v)
}

type transitionReq struct {
	ExpectedState string `json:"expected_state"`
	NextState     string `json:"next_state"`
	ActorType     string `json:"actor_type"`
}

func (h *Handler) TransitionVerification(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 { // /api/v1/verifications/{id}/transition
		writeError(w, http.StatusBadRequest, domain.ErrInvalidRequest)
		return
	}
	id := pathParts[4]

	var req transitionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, domain.ErrInvalidRequest)
		return
	}

	err := h.svc.Transition(r.Context(), id, domain.VerificationStatus(req.ExpectedState), domain.VerificationStatus(req.NextState), req.ActorType, "")
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrInvalidStateTransition {
			status = http.StatusConflict
		}
		writeError(w, status, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    err.Error(),
			"message": err.Error(),
		},
	})
}
