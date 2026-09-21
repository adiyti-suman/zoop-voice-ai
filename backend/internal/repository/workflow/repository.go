package workflow

import (
	"context"

	wf "verification-platform/internal/domain/workflow"
)

// Repository is the persistence contract for the workflow domain.
type Repository interface {
	// Workflow CRUD
	CreateWorkflow(ctx context.Context, w *wf.Workflow) error
	GetWorkflow(ctx context.Context, id string) (*wf.Workflow, error)

	// Version CRUD
	CreateVersion(ctx context.Context, v *wf.WorkflowVersion) error
	GetVersion(ctx context.Context, id string) (*wf.WorkflowVersion, error)
	UpdateVersionStatus(ctx context.Context, id string, status wf.WorkflowVersionStatus) error

	// Graph
	UpsertNodes(ctx context.Context, nodes []*wf.WorkflowNode) error
	UpsertEdges(ctx context.Context, edges []*wf.WorkflowEdge) error
	GetNodes(ctx context.Context, versionID string) ([]*wf.WorkflowNode, error)
	GetEdges(ctx context.Context, versionID string) ([]*wf.WorkflowEdge, error)

	// Runs
	CreateRun(ctx context.Context, r *wf.WorkflowRun) error
	GetRun(ctx context.Context, id string) (*wf.WorkflowRun, error)
	UpdateRunStatus(ctx context.Context, id string, status wf.WorkflowRunStatus, errPayload *wf.RunError) error
	UpdateRunCurrentNode(ctx context.Context, id string, nodeKey string) error
	SetRunOutput(ctx context.Context, id string, output map[string]interface{}) error

	// Node Runs
	CreateNodeRun(ctx context.Context, nr *wf.NodeRun) error
	UpdateNodeRun(ctx context.Context, nr *wf.NodeRun) error
	GetNodeRuns(ctx context.Context, runID string) ([]*wf.NodeRun, error)
}
