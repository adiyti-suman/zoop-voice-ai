package voice

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type SessionStatus string

const (
	SessionStatusCreated    SessionStatus = "CREATED"
	SessionStatusConnecting SessionStatus = "CONNECTING"
	SessionStatusConnected  SessionStatus = "CONNECTED"
	SessionStatusListening  SessionStatus = "LISTENING"
	SessionStatusPaused     SessionStatus = "PAUSED"
	SessionStatusStopping   SessionStatus = "STOPPING"
	SessionStatusCompleted  SessionStatus = "COMPLETED"
	SessionStatusFailed     SessionStatus = "FAILED"
	SessionStatusCancelled  SessionStatus = "CANCELLED"
)

var (
	ErrSessionNotFound     = errors.New("VOICE_SESSION_NOT_FOUND")
	ErrInvalidTransition   = errors.New("INVALID_SESSION_TRANSITION")
	ErrInvalidAudio        = errors.New("INVALID_AUDIO")
	ErrBufferOverflow      = errors.New("AUDIO_BUFFER_OVERFLOW")
	ErrSequenceInvalid     = errors.New("SEQUENCE_INVALID")
)

type VoiceSession struct {
	ID             uuid.UUID
	VerificationID uuid.UUID
	WorkflowRunID  uuid.UUID
	Status         SessionStatus
	Language       *string
	SampleRate     int
	Channels       int
	Encoding       string
	StartedAt      *time.Time
	EndedAt        *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IsTerminal returns true if the session is in a final state.
func IsTerminal(s SessionStatus) bool {
	return s == SessionStatusCompleted || s == SessionStatusFailed || s == SessionStatusCancelled
}

// IsTransitionValid enforces the state machine for voice sessions.
func IsTransitionValid(from, to SessionStatus) bool {
	// Cancel/Fail are broadly reachable
	if to == SessionStatusCancelled {
		return from == SessionStatusCreated || from == SessionStatusConnected || from == SessionStatusListening || from == SessionStatusPaused
	}
	if to == SessionStatusFailed {
		return from == SessionStatusConnecting || from == SessionStatusConnected || from == SessionStatusListening
	}

	switch from {
	case SessionStatusCreated:
		return to == SessionStatusConnecting
	case SessionStatusConnecting:
		return to == SessionStatusConnected
	case SessionStatusConnected:
		return to == SessionStatusListening || to == SessionStatusStopping
	case SessionStatusListening:
		return to == SessionStatusPaused || to == SessionStatusStopping
	case SessionStatusPaused:
		return to == SessionStatusListening || to == SessionStatusStopping
	case SessionStatusStopping:
		return to == SessionStatusCompleted
	}
	return false
}
