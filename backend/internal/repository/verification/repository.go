package verification

import (
	"context"

	"verification-platform/internal/domain/verification"
)

type Repository interface {
	Create(ctx context.Context, v *verification.Verification) error
	CreateIdempotent(ctx context.Context, v *verification.Verification, idempotencyKey string) (*verification.Verification, error)
	GetByID(ctx context.Context, id string) (*verification.Verification, error)
	GetByExternalID(ctx context.Context, externalID string) (*verification.Verification, error)
	UpdateStatus(ctx context.Context, id string, expectedStatus verification.VerificationStatus, newStatus verification.VerificationStatus) error
	Update(ctx context.Context, v *verification.Verification) error
	LogAuditEvent(ctx context.Context, event *verification.AuditEvent) error
}
