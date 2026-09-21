package voice

import (
	"context"
	"time"

	"github.com/google/uuid"

	domain "verification-platform/internal/domain/voice"
	repo "verification-platform/internal/repository/voice"
)

type CreateSessionRequest struct {
	VerificationID string `json:"verification_id"`
	WorkflowRunID  string `json:"workflow_run_id"`
	SampleRate     int    `json:"sample_rate"`
	Channels       int    `json:"channels"`
	Encoding       string `json:"encoding"`
}

type Service interface {
	CreateSession(ctx context.Context, req CreateSessionRequest) (*domain.VoiceSession, error)
	GetSession(ctx context.Context, id string) (*domain.VoiceSession, error)
	TransitionStatus(ctx context.Context, id string, newStatus domain.SessionStatus) error
}

type serviceImpl struct {
	repo repo.Repository
}

func NewService(repo repo.Repository) Service {
	return &serviceImpl{repo: repo}
}

func (s *serviceImpl) CreateSession(ctx context.Context, req CreateSessionRequest) (*domain.VoiceSession, error) {
	verUUID, err := uuid.Parse(req.VerificationID)
	if err != nil {
		return nil, err
	}
	wfUUID, err := uuid.Parse(req.WorkflowRunID)
	if err != nil {
		return nil, err
	}

	// Validate Canonical Audio Format
	if req.SampleRate != 16000 || req.Channels != 1 || req.Encoding != "pcm_s16le" {
		return nil, domain.ErrInvalidAudio
	}

	now := time.Now().UTC()
	session := &domain.VoiceSession{
		ID:             uuid.New(),
		VerificationID: verUUID,
		WorkflowRunID:  wfUUID,
		Status:         domain.SessionStatusCreated,
		SampleRate:     req.SampleRate,
		Channels:       req.Channels,
		Encoding:       req.Encoding,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *serviceImpl) GetSession(ctx context.Context, id string) (*domain.VoiceSession, error) {
	return s.repo.GetSession(ctx, id)
}

func (s *serviceImpl) TransitionStatus(ctx context.Context, id string, newStatus domain.SessionStatus) error {
	session, err := s.repo.GetSession(ctx, id)
	if err != nil {
		return err
	}

	if !domain.IsTransitionValid(session.Status, newStatus) {
		return domain.ErrInvalidTransition
	}

	return s.repo.UpdateSessionStatus(ctx, id, newStatus)
}
