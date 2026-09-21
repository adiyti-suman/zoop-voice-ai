package workflow

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WorkflowRunStatus is the lifecycle state of a workflow execution.
type WorkflowRunStatus string

const (
	RunStatusPending   WorkflowRunStatus = "PENDING"
	RunStatusRunning   WorkflowRunStatus = "RUNNING"
	RunStatusPaused    WorkflowRunStatus = "PAUSED"
	RunStatusWaiting   WorkflowRunStatus = "WAITING"
	RunStatusCompleted WorkflowRunStatus = "COMPLETED"
	RunStatusFailed    WorkflowRunStatus = "FAILED"
	RunStatusCancelled WorkflowRunStatus = "CANCELLED"
	RunStatusTimedOut  WorkflowRunStatus = "TIMED_OUT"
)

// IsRunTerminal returns true when no further transitions are possible.
func IsRunTerminal(s WorkflowRunStatus) bool {
	switch s {
	case RunStatusCompleted, RunStatusFailed, RunStatusCancelled, RunStatusTimedOut:
		return true
	}
	return false
}

// IsRunTransitionValid returns true if the transition is permitted.
func IsRunTransitionValid(from, to WorkflowRunStatus) bool {
	switch from {
	case RunStatusPending:
		return to == RunStatusRunning || to == RunStatusCancelled
	case RunStatusRunning:
		return to == RunStatusCompleted || to == RunStatusFailed ||
			to == RunStatusCancelled || to == RunStatusPaused ||
			to == RunStatusWaiting || to == RunStatusTimedOut
	case RunStatusWaiting:
		return to == RunStatusRunning || to == RunStatusCancelled
	case RunStatusPaused:
		return to == RunStatusRunning || to == RunStatusCancelled
	}
	return false
}

// NodeRunStatus is the lifecycle state of a single node execution attempt.
type NodeRunStatus string

const (
	NodeRunPending   NodeRunStatus = "PENDING"
	NodeRunRunning   NodeRunStatus = "RUNNING"
	NodeRunWaiting   NodeRunStatus = "WAITING"
	NodeRunCompleted NodeRunStatus = "COMPLETED"
	NodeRunFailed    NodeRunStatus = "FAILED"
	NodeRunSkipped   NodeRunStatus = "SKIPPED"
	NodeRunCancelled NodeRunStatus = "CANCELLED"
	NodeRunTimedOut  NodeRunStatus = "TIMED_OUT"
)

// WorkflowRun represents one execution of a published workflow version.
type WorkflowRun struct {
	ID                uuid.UUID              `json:"id"`
	VerificationID    uuid.UUID              `json:"verification_id"`
	WorkflowVersionID uuid.UUID              `json:"workflow_version_id"`
	Status            WorkflowRunStatus      `json:"status"`
	CurrentNodeKey    *string                `json:"current_node_key,omitempty"`
	Input             map[string]interface{} `json:"input"`
	Output            map[string]interface{} `json:"output,omitempty"`
	Error             *RunError              `json:"error,omitempty"`
	StartedAt         *time.Time             `json:"started_at,omitempty"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// RunError carries structured failure information.
type RunError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	NodeKey string `json:"node_key,omitempty"`
}

func (r *WorkflowRun) InputBytes() ([]byte, error) {
	if r.Input == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(r.Input)
}

func (r *WorkflowRun) OutputBytes() ([]byte, error) {
	if r.Output == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(r.Output)
}

func (r *WorkflowRun) ErrorBytes() ([]byte, error) {
	if r.Error == nil {
		return []byte("null"), nil
	}
	return json.Marshal(r.Error)
}

// NodeRun records one execution attempt of one node within a workflow run.
type NodeRun struct {
	ID            uuid.UUID              `json:"id"`
	WorkflowRunID uuid.UUID              `json:"workflow_run_id"`
	NodeKey       string                 `json:"node_key"`
	Attempt       int                    `json:"attempt"`
	Status        NodeRunStatus          `json:"status"`
	Input         map[string]interface{} `json:"input,omitempty"`
	Output        map[string]interface{} `json:"output,omitempty"`
	Error         *RunError              `json:"error,omitempty"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	DurationMs    int64                  `json:"duration_ms,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

func (n *NodeRun) InputBytes() ([]byte, error) {
	if n.Input == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(n.Input)
}

func (n *NodeRun) OutputBytes() ([]byte, error) {
	if n.Output == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(n.Output)
}
