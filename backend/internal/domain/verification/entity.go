package verification

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Verification struct {
	ID              uuid.UUID              `json:"id"`
	ExternalID      *string                `json:"external_id,omitempty"`
	Type            string                 `json:"type"`
	Status          VerificationStatus     `json:"status"`
	SubjectID       *string                `json:"subject_id,omitempty"`
	WorkflowID      *uuid.UUID             `json:"workflow_id,omitempty"`
	WorkflowVersion *string                `json:"workflow_version,omitempty"`
	Language        *string                `json:"language,omitempty"`
	Locale          *string                `json:"locale,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	StartedAt       *time.Time             `json:"started_at,omitempty"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
	ExpiresAt       *time.Time             `json:"expires_at,omitempty"`
}

type AuditEvent struct {
	ID             uuid.UUID              `json:"id"`
	VerificationID uuid.UUID              `json:"verification_id"`
	EventType      string                 `json:"event_type"`
	ActorType      string                 `json:"actor_type"`
	ActorID        *string                `json:"actor_id,omitempty"`
	PreviousState  *VerificationStatus    `json:"previous_state,omitempty"`
	NewState       *VerificationStatus    `json:"new_state,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
}

// Validate validates the core requirements of a Verification
func (v *Verification) Validate() error {
	if v.Type == "" {
		return ErrInvalidVerificationType
	}
	return nil
}

// EnsureMetadata ensures the metadata map is not nil
func (v *Verification) EnsureMetadata() {
	if v.Metadata == nil {
		v.Metadata = make(map[string]interface{})
	}
}

// GetMetadataBytes returns the metadata as JSON bytes
func (v *Verification) GetMetadataBytes() ([]byte, error) {
	v.EnsureMetadata()
	return json.Marshal(v.Metadata)
}

// SetMetadataBytes unmarshals metadata from JSON bytes
func (v *Verification) SetMetadataBytes(data []byte) error {
	if len(data) == 0 {
		v.Metadata = make(map[string]interface{})
		return nil
	}
	return json.Unmarshal(data, &v.Metadata)
}
