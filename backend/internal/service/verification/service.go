package verification

import (
	"context"
	"time"

	"github.com/google/uuid"
	"verification-platform/internal/domain/verification"
	verfRepo "verification-platform/internal/repository/verification"
)

type CreateRequest struct {
	Type            string                 `json:"type"`
	SubjectID       *string                `json:"subject_id,omitempty"`
	WorkflowID      *string                `json:"workflow_id,omitempty"`
	WorkflowVersion *string                `json:"workflow_version,omitempty"`
	Language        *string                `json:"language,omitempty"`
	Locale          *string                `json:"locale,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type Service interface {
	Create(ctx context.Context, idempotencyKey string, req CreateRequest) (*verification.Verification, error)
	Get(ctx context.Context, id string) (*verification.Verification, error)
	Transition(ctx context.Context, id string, expected verification.VerificationStatus, next verification.VerificationStatus, actorType, actorID string) error
}

type serviceImpl struct {
	repo verfRepo.Repository
}

func NewService(repo verfRepo.Repository) Service {
	return &serviceImpl{repo: repo}
}

func (s *serviceImpl) Create(ctx context.Context, idempotencyKey string, req CreateRequest) (*verification.Verification, error) {
	// Simple Idempotency logic - store in DB as external ID for this phase
	// Or use Redis properly. For Phase 03 we use external_id as a fallback.
	if idempotencyKey != "" {
		existing, err := s.repo.GetByExternalID(ctx, idempotencyKey)
		if err == nil {
			return existing, nil
		}
	}

	var wfID *uuid.UUID
	if req.WorkflowID != nil && *req.WorkflowID != "" {
		parsed, err := uuid.Parse(*req.WorkflowID)
		if err == nil {
			wfID = &parsed
		}
	}

	extID := &idempotencyKey
	if idempotencyKey == "" {
		extID = nil
	}

	now := time.Now().UTC()
	v := &verification.Verification{
		ID:              uuid.New(),
		ExternalID:      extID,
		Type:            req.Type,
		Status:          verification.StatusCreated,
		SubjectID:       req.SubjectID,
		WorkflowID:      wfID,
		WorkflowVersion: req.WorkflowVersion,
		Language:        req.Language,
		Locale:          req.Locale,
		Metadata:        req.Metadata,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := v.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, v); err != nil {
		return nil, verification.ErrInternalError
	}

	event := &verification.AuditEvent{
		ID:             uuid.New(),
		VerificationID: v.ID,
		EventType:      "VERIFICATION_CREATED",
		ActorType:      "SYSTEM",
		NewState:       &v.Status,
		Metadata:       map[string]interface{}{},
		CreatedAt:      now,
	}
	_ = s.repo.LogAuditEvent(ctx, event) // Ignore audit log failure for now

	return v, nil
}

func (s *serviceImpl) Get(ctx context.Context, id string) (*verification.Verification, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, verification.ErrInvalidRequest
	}
	return s.repo.GetByID(ctx, id)
}

func (s *serviceImpl) Transition(ctx context.Context, id string, expected verification.VerificationStatus, next verification.VerificationStatus, actorType, actorID string) error {
	if !verification.IsValidTransition(expected, next) {
		return verification.ErrInvalidStateTransition
	}

	err := s.repo.UpdateStatus(ctx, id, expected, next)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	parsedID, _ := uuid.Parse(id)
	var aID *string
	if actorID != "" {
		aID = &actorID
	}
	
	event := &verification.AuditEvent{
		ID:             uuid.New(),
		VerificationID: parsedID,
		EventType:      "VERIFICATION_STATUS_CHANGED",
		ActorType:      actorType,
		ActorID:        aID,
		PreviousState:  &expected,
		NewState:       &next,
		Metadata:       map[string]interface{}{},
		CreatedAt:      now,
	}
	_ = s.repo.LogAuditEvent(ctx, event)

	return nil
}
