package workflow

import (
	"time"

	"github.com/google/uuid"
)

// WorkflowStatus is the lifecycle state of a workflow definition.
type WorkflowStatus string

const (
	WorkflowStatusDraft    WorkflowStatus = "DRAFT"
	WorkflowStatusActive   WorkflowStatus = "ACTIVE"
	WorkflowStatusArchived WorkflowStatus = "ARCHIVED"
)

// WorkflowVersionStatus is the lifecycle state of a workflow version.
type WorkflowVersionStatus string

const (
	VersionStatusDraft      WorkflowVersionStatus = "DRAFT"
	VersionStatusValidating WorkflowVersionStatus = "VALIDATING"
	VersionStatusPublished  WorkflowVersionStatus = "PUBLISHED"
	VersionStatusArchived   WorkflowVersionStatus = "ARCHIVED"
)

// IsVersionTransitionValid returns true if the transition is allowed.
func IsVersionTransitionValid(from, to WorkflowVersionStatus) bool {
	switch from {
	case VersionStatusDraft:
		return to == VersionStatusValidating || to == VersionStatusPublished || to == VersionStatusArchived
	case VersionStatusValidating:
		return to == VersionStatusPublished || to == VersionStatusDraft
	case VersionStatusPublished:
		return to == VersionStatusArchived
	}
	return false
}

// Workflow is the top-level logical definition of a verification process.
type Workflow struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      WorkflowStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// WorkflowVersion is an immutable, versioned snapshot of a Workflow's graph.
type WorkflowVersion struct {
	ID          uuid.UUID             `json:"id"`
	WorkflowID  uuid.UUID             `json:"workflow_id"`
	Version     string                `json:"version"`
	Status      WorkflowVersionStatus `json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
	PublishedAt *time.Time            `json:"published_at,omitempty"`
	ArchivedAt  *time.Time            `json:"archived_at,omitempty"`
}
