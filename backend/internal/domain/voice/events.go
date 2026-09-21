package voice

import (
	"time"
)

// VoiceEvent is an internal event interface for telemetry and future observers (Phase 06, 07)
type VoiceEvent struct {
	Type      string         `json:"type"`
	SessionID string         `json:"session_id"`
	Timestamp time.Time      `json:"timestamp"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

const (
	EventSessionCreated   = "SESSION_CREATED"
	EventSessionConnected = "SESSION_CONNECTED"
	EventSessionListening = "SESSION_LISTENING"
	EventSessionPaused    = "SESSION_PAUSED"
	EventSessionResumed   = "SESSION_RESUMED"
	EventSessionStopped   = "SESSION_STOPPED"
	EventSessionCompleted = "SESSION_COMPLETED"
	EventSessionFailed    = "SESSION_FAILED"
	EventSessionCancelled = "SESSION_CANCELLED"

	EventAudioChunkReceived = "AUDIO_CHUNK_RECEIVED"
	EventAudioDropped       = "AUDIO_CHUNK_DROPPED"
)
