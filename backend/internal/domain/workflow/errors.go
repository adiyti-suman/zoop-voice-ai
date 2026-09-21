package workflow

import "errors"

var (
	ErrWorkflowNotFound     = errors.New("WORKFLOW_NOT_FOUND")
	ErrVersionNotFound      = errors.New("WORKFLOW_VERSION_NOT_FOUND")
	ErrVersionNotPublished  = errors.New("WORKFLOW_VERSION_NOT_PUBLISHED")
	ErrRunNotFound          = errors.New("WORKFLOW_RUN_NOT_FOUND")
	ErrCycleDetected        = errors.New("WORKFLOW_CYCLE_DETECTED")
	ErrInvalidGraph         = errors.New("INVALID_WORKFLOW_GRAPH")
	ErrNodeTypeNotSupported = errors.New("NODE_TYPE_NOT_SUPPORTED")
	ErrInvalidTransition    = errors.New("INVALID_WORKFLOW_STATUS_TRANSITION")
	ErrRunAlreadyTerminal   = errors.New("WORKFLOW_RUN_ALREADY_TERMINAL")
	ErrInvalidRequest       = errors.New("INVALID_REQUEST")
)
