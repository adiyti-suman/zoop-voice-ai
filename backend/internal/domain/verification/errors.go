package verification

import (
	"errors"
)

var (
	ErrVerificationNotFound     = errors.New("VERIFICATION_NOT_FOUND")
	ErrInvalidVerificationType  = errors.New("INVALID_VERIFICATION_TYPE")
	ErrInvalidStateTransition   = errors.New("INVALID_STATE_TRANSITION")
	ErrInvalidRequest           = errors.New("INVALID_REQUEST")
	ErrDuplicateIdempotencyKey  = errors.New("DUPLICATE_IDEMPOTENCY_KEY")
	ErrVerificationExpired      = errors.New("VERIFICATION_EXPIRED")
	ErrInternalError            = errors.New("INTERNAL_ERROR")
)
