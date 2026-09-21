package workflow

import (
	"context"
	"time"

	"github.com/google/uuid"

	wf "verification-platform/internal/domain/workflow"
	engine "verification-platform/internal/engine/workflow"
	"verification-platform/internal/engine/workflow/executor"
	repo "verification-platform/internal/repository/workflow"
)

// CreateWorkflowRequest is the input for creating a new workflow definition.
type CreateWorkflowRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateVersionRequest is the input for creating a new workflow version.
type CreateVersionRequest struct {
	Version string      `json:"version"`
	Nodes   []NodeInput `json:"nodes"`
	Edges   []EdgeInput `json:"edges"`
}

// NodeInput is the API representation of a workflow node.
type NodeInput struct {
	NodeKey     string                 `json:"node_key"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Config      map[string]interface{} `json:"config"`
	Position    wf.NodePosition        `json:"position"`
	TimeoutMs   int                    `json:"timeout_ms"`
	RetryPolicy *wf.RetryPolicy        `json:"retry_policy,omitempty"`
}

// EdgeInput is the API representation of a workflow edge.
type EdgeInput struct {
	SourceNodeKey string                 `json:"source_node_key"`
	TargetNodeKey string                 `json:"target_node_key"`
	Condition     string                 `json:"condition"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// StartRunRequest is the input for starting a workflow run.
type StartRunRequest struct {
	VerificationID string                 `json:"verification_id"`
	Input          map[string]interface{} `json:"input"`
}

// Service is the application-level workflow service.
type Service interface {
	CreateWorkflow(ctx context.Context, req CreateWorkflowRequest) (*wf.Workflow, error)
	GetWorkflow(ctx context.Context, id string) (*wf.Workflow, error)
	CreateVersion(ctx context.Context, workflowID string, req CreateVersionRequest) (*wf.WorkflowVersion, error)
	GetVersion(ctx context.Context, id string) (*wf.WorkflowVersion, error)
	ValidateVersion(ctx context.Context, id string) (engine.ValidationResult, error)
	PublishVersion(ctx context.Context, id string) (*wf.WorkflowVersion, error)
	StartRun(ctx context.Context, versionID string, req StartRunRequest) (*wf.WorkflowRun, error)
	GetRun(ctx context.Context, id string) (*wf.WorkflowRun, []*wf.NodeRun, error)
	CancelRun(ctx context.Context, id string) error
}

type serviceImpl struct {
	repo   repo.Repository
	engine *engine.Engine
}

// NewService builds the workflow service and registers all built-in executors.
func NewService(r repo.Repository) Service {
	registry := engine.NewRegistry()
	registry.Register(&executor.StartExecutor{})
	registry.Register(&executor.EndExecutor{})
	registry.Register(&executor.NoopExecutor{})
	registry.Register(&executor.TransformExecutor{})
	registry.Register(&executor.ConditionExecutor{})

	eng := engine.NewEngine(registry, r)
	return &serviceImpl{repo: r, engine: eng}
}

func (s *serviceImpl) CreateWorkflow(ctx context.Context, req CreateWorkflowRequest) (*wf.Workflow, error) {
	if req.Name == "" {
		return nil, wf.ErrInvalidRequest
	}
	now := time.Now().UTC()
	w := &wf.Workflow{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Status:      wf.WorkflowStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateWorkflow(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *serviceImpl) GetWorkflow(ctx context.Context, id string) (*wf.Workflow, error) {
	return s.repo.GetWorkflow(ctx, id)
}

func (s *serviceImpl) CreateVersion(ctx context.Context, workflowID string, req CreateVersionRequest) (*wf.WorkflowVersion, error) {
	if req.Version == "" {
		return nil, wf.ErrInvalidRequest
	}
	wfID, err := uuid.Parse(workflowID)
	if err != nil {
		return nil, wf.ErrInvalidRequest
	}

	now := time.Now().UTC()
	v := &wf.WorkflowVersion{
		ID:         uuid.New(),
		WorkflowID: wfID,
		Version:    req.Version,
		Status:     wf.VersionStatusDraft,
		CreatedAt:  now,
	}
	if err := s.repo.CreateVersion(ctx, v); err != nil {
		return nil, err
	}

	// Persist nodes
	var nodes []*wf.WorkflowNode
	for _, n := range req.Nodes {
		nodes = append(nodes, &wf.WorkflowNode{
			ID:                uuid.New(),
			WorkflowVersionID: v.ID,
			NodeKey:           n.NodeKey,
			Type:              wf.NodeType(n.Type),
			Name:              n.Name,
			Config:            n.Config,
			Position:          n.Position,
			TimeoutMs:         n.TimeoutMs,
			RetryPolicy:       n.RetryPolicy,
			CreatedAt:         now,
		})
	}
	if len(nodes) > 0 {
		if err := s.repo.UpsertNodes(ctx, nodes); err != nil {
			return nil, err
		}
	}

	// Persist edges
	var edges []*wf.WorkflowEdge
	for _, e := range req.Edges {
		edges = append(edges, &wf.WorkflowEdge{
			ID:                uuid.New(),
			WorkflowVersionID: v.ID,
			SourceNodeKey:     e.SourceNodeKey,
			TargetNodeKey:     e.TargetNodeKey,
			Condition:         e.Condition,
			Metadata:          e.Metadata,
			CreatedAt:         now,
		})
	}
	if len(edges) > 0 {
		if err := s.repo.UpsertEdges(ctx, edges); err != nil {
			return nil, err
		}
	}

	return v, nil
}

func (s *serviceImpl) GetVersion(ctx context.Context, id string) (*wf.WorkflowVersion, error) {
	return s.repo.GetVersion(ctx, id)
}

func (s *serviceImpl) ValidateVersion(ctx context.Context, id string) (engine.ValidationResult, error) {
	nodes, err := s.repo.GetNodes(ctx, id)
	if err != nil {
		return engine.ValidationResult{}, err
	}
	edges, err := s.repo.GetEdges(ctx, id)
	if err != nil {
		return engine.ValidationResult{}, err
	}
	return engine.Validate(nodes, edges), nil
}

func (s *serviceImpl) PublishVersion(ctx context.Context, id string) (*wf.WorkflowVersion, error) {
	v, err := s.repo.GetVersion(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate first
	result, err := s.ValidateVersion(ctx, id)
	if err != nil {
		return nil, err
	}
	if !result.Valid {
		return nil, wf.ErrInvalidGraph
	}

	if !wf.IsVersionTransitionValid(v.Status, wf.VersionStatusPublished) {
		return nil, wf.ErrInvalidTransition
	}

	if err := s.repo.UpdateVersionStatus(ctx, id, wf.VersionStatusPublished); err != nil {
		return nil, err
	}
	v.Status = wf.VersionStatusPublished
	return v, nil
}

func (s *serviceImpl) StartRun(ctx context.Context, versionID string, req StartRunRequest) (*wf.WorkflowRun, error) {
	v, err := s.repo.GetVersion(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if v.Status != wf.VersionStatusPublished {
		return nil, wf.ErrVersionNotPublished
	}

	verificationID, err := uuid.Parse(req.VerificationID)
	if err != nil {
		return nil, wf.ErrInvalidRequest
	}

	now := time.Now().UTC()
	run := &wf.WorkflowRun{
		ID:                uuid.New(),
		VerificationID:    verificationID,
		WorkflowVersionID: v.ID,
		Status:            wf.RunStatusPending,
		Input:             req.Input,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if run.Input == nil {
		run.Input = map[string]interface{}{}
	}

	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, err
	}

	// Execute synchronously (Phase 04 — no AI nodes, fast execution)
	if err := s.engine.Run(ctx, run.ID.String()); err != nil {
		// Engine failure is recorded in the DB; return the run for inspection
		_ = err
	}

	// Return the updated run
	updatedRun, err := s.repo.GetRun(ctx, run.ID.String())
	if err != nil {
		return run, nil // Best effort
	}
	return updatedRun, nil
}

func (s *serviceImpl) GetRun(ctx context.Context, id string) (*wf.WorkflowRun, []*wf.NodeRun, error) {
	run, err := s.repo.GetRun(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	nodeRuns, err := s.repo.GetNodeRuns(ctx, id)
	if err != nil {
		return run, nil, nil
	}
	return run, nodeRuns, nil
}

func (s *serviceImpl) CancelRun(ctx context.Context, id string) error {
	run, err := s.repo.GetRun(ctx, id)
	if err != nil {
		return err
	}
	if wf.IsRunTerminal(run.Status) {
		return wf.ErrRunAlreadyTerminal
	}
	return s.repo.UpdateRunStatus(ctx, id, wf.RunStatusCancelled, nil)
}
