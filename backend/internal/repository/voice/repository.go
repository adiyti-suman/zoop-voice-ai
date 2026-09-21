package voice

import (
	"context"

	domain "verification-platform/internal/domain/voice"
)

// Repository manages persistence for VoiceSessions.
type Repository interface {
	CreateSession(ctx context.Context, session *domain.VoiceSession) error
	GetSession(ctx context.Context, id string) (*domain.VoiceSession, error)
	UpdateSessionStatus(ctx context.Context, id string, status domain.SessionStatus) error
}
