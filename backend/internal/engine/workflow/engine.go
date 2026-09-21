package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	wf "verification-platform/internal/domain/workflow"
	repo "verification-platform/internal/repository/workflow"
)

// Engine orchestrates the execution of a WorkflowRun.
type Engine struct {
	registry *Registry
	repo     repo.Repository
}

func NewEngine(registry *Registry, repo repo.Repository) *Engine {
	return &Engine{registry: registry, repo: repo}
}

// Run executes a workflow run synchronously.
// It loads the graph, iterates nodes from START, and persists every step.
func (e *Engine) Run(ctx context.Context, runID string) error {
	run, err := e.repo.GetRun(ctx, runID)
	if err != nil {
		return err
	}

	if wf.IsRunTerminal(run.Status) {
		return wf.ErrRunAlreadyTerminal
	}

	// Mark run as RUNNING
	if err := e.repo.UpdateRunStatus(ctx, runID, wf.RunStatusRunning, nil); err != nil {
		return err
	}

	// Load graph
	nodes, err := e.repo.GetNodes(ctx, run.WorkflowVersionID.String())
	if err != nil {
		return e.failRun(ctx, runID, "GRAPH_LOAD_ERROR", err.Error(), "")
	}
	edges, err := e.repo.GetEdges(ctx, run.WorkflowVersionID.String())
	if err != nil {
		return e.failRun(ctx, runID, "GRAPH_LOAD_ERROR", err.Error(), "")
	}

	// Build lookup structures
	nodeByKey := map[string]*wf.WorkflowNode{}
	for _, n := range nodes {
		nodeByKey[n.NodeKey] = n
	}
	// adjacency: source → edges
	adjEdges := map[string][]*wf.WorkflowEdge{}
	for _, edge := range edges {
		adjEdges[edge.SourceNodeKey] = append(adjEdges[edge.SourceNodeKey], edge)
	}

	// Find START node
	var startKey string
	for _, n := range nodes {
		if n.Type == wf.NodeTypeStart {
			startKey = n.NodeKey
			break
		}
	}
	if startKey == "" {
		return e.failRun(ctx, runID, "NO_START_NODE", "no START node found", "")
	}

	// Build execution context
	execContext := map[string]interface{}{
		"verification_id":    run.VerificationID.String(),
		"workflow_version_id": run.WorkflowVersionID.String(),
		"run_id":              runID,
	}
	for k, v := range run.Input {
		execContext[k] = v
	}

	nodeOutputs := map[string]map[string]interface{}{}

	// Execute nodes sequentially
	currentKey := startKey
	visited := map[string]bool{}

	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return e.failRun(ctx, runID, "RUN_CANCELLED", "context cancelled", currentKey)
		default:
		}

		if visited[currentKey] {
			return e.failRun(ctx, runID, "CYCLE_DETECTED", fmt.Sprintf("node %s visited twice", currentKey), currentKey)
		}
		visited[currentKey] = true

		node, ok := nodeByKey[currentKey]
		if !ok {
			return e.failRun(ctx, runID, "NODE_NOT_FOUND", fmt.Sprintf("node %s not found", currentKey), currentKey)
		}

		// Update current node in DB
		e.repo.UpdateRunCurrentNode(ctx, runID, currentKey)

		// Execute the node (with retry if configured)
		output, nodeRunErr := e.executeNodeWithRetry(ctx, run, node, execContext, nodeOutputs)

		if nodeRunErr != nil {
			return e.failRun(ctx, runID, "NODE_EXECUTION_FAILED", nodeRunErr.Error(), currentKey)
		}

		// Merge output into shared context
		nodeOutputs[currentKey] = output.Data

		// If END, we're done
		if node.Type == wf.NodeTypeEnd {
			// flatten nodeOutputs into map[string]interface{} for storage
			flat := map[string]interface{}{}
			for k, v := range nodeOutputs {
				flat[k] = v
			}
			e.repo.SetRunOutput(ctx, runID, flat)
			return e.repo.UpdateRunStatus(ctx, runID, wf.RunStatusCompleted, nil)
		}

		// Determine next node
		outEdges := adjEdges[currentKey]
		if len(outEdges) == 0 {
			return e.failRun(ctx, runID, "NO_OUTGOING_EDGE", fmt.Sprintf("node %s has no outgoing edges", currentKey), currentKey)
		}

		nextKey, err := e.selectNextNode(outEdges, output, execContext, nodeOutputs)
		if err != nil {
			return e.failRun(ctx, runID, "EDGE_SELECTION_ERROR", err.Error(), currentKey)
		}
		currentKey = nextKey
	}
}

// executeNodeWithRetry runs a node, retrying on failure per the node's RetryPolicy.
func (e *Engine) executeNodeWithRetry(
	ctx context.Context,
	run *wf.WorkflowRun,
	node *wf.WorkflowNode,
	execCtx map[string]interface{},
	nodeOutputs map[string]map[string]interface{},
) (ExecutionOutput, error) {
	maxAttempts := 1
	backoffMs := 0
	if node.RetryPolicy != nil && node.RetryPolicy.MaxAttempts > 0 {
		maxAttempts = node.RetryPolicy.MaxAttempts
		backoffMs = node.RetryPolicy.BackoffMs
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		output, err := e.executeNode(ctx, run, node, execCtx, nodeOutputs, attempt)
		if err == nil {
			return output, nil
		}
		lastErr = err
		if attempt < maxAttempts && backoffMs > 0 {
			select {
			case <-ctx.Done():
				return ExecutionOutput{}, ctx.Err()
			case <-time.After(time.Duration(backoffMs*attempt) * time.Millisecond):
			}
		}
	}
	return ExecutionOutput{}, lastErr
}

// executeNode runs a single attempt of a node, persisting a NodeRun record.
func (e *Engine) executeNode(
	ctx context.Context,
	run *wf.WorkflowRun,
	node *wf.WorkflowNode,
	execCtx map[string]interface{},
	nodeOutputs map[string]map[string]interface{},
	attempt int,
) (ExecutionOutput, error) {
	executor, ok := e.registry.Get(string(node.Type))
	if !ok {
		return ExecutionOutput{}, fmt.Errorf("no executor registered for node type: %s", node.Type)
	}

	now := time.Now().UTC()
	nr := &wf.NodeRun{
		ID:            uuid.New(),
		WorkflowRunID: run.ID,
		NodeKey:       node.NodeKey,
		Attempt:       attempt,
		Status:        wf.NodeRunRunning,
		StartedAt:     &now,
		Input:         execCtx,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	e.repo.CreateNodeRun(ctx, nr)

	// Apply timeout if configured
	execCtxRun := ctx
	var cancel context.CancelFunc
	if node.TimeoutMs > 0 {
		execCtxRun, cancel = context.WithTimeout(ctx, time.Duration(node.TimeoutMs)*time.Millisecond)
		defer cancel()
	}

	input := ExecutionInput{
		NodeKey:     node.NodeKey,
		RunID:       run.ID.String(),
		NodeConfig:  node.Config,
		Context:     execCtx,
		NodeOutputs: nodeOutputs,
	}

	start := time.Now()
	output, err := executor.Execute(execCtxRun, input)
	duration := time.Since(start).Milliseconds()
	completedAt := time.Now().UTC()
	nr.DurationMs = duration
	nr.CompletedAt = &completedAt
	nr.UpdatedAt = completedAt

	if err != nil {
		nr.Status = wf.NodeRunFailed
		nr.Error = &wf.RunError{Code: "EXECUTOR_ERROR", Message: err.Error(), NodeKey: node.NodeKey}
	} else if execCtxRun.Err() != nil {
		nr.Status = wf.NodeRunTimedOut
		nr.Error = &wf.RunError{Code: "TIMED_OUT", Message: "node execution timed out", NodeKey: node.NodeKey}
		err = fmt.Errorf("node %s timed out", node.NodeKey)
	} else {
		nr.Status = wf.NodeRunCompleted
		nr.Output = output.Data
	}

	e.repo.UpdateNodeRun(ctx, nr)
	return output, err
}

// selectNextNode picks the next node key from the available edges.
// For CONDITION nodes, it matches output.NextEdgeLabel against edge conditions.
// For all others, it picks the first unconditional edge.
func (e *Engine) selectNextNode(
	outEdges []*wf.WorkflowEdge,
	output ExecutionOutput,
	execCtx map[string]interface{},
	nodeOutputs map[string]map[string]interface{},
) (string, error) {
	if output.NextEdgeLabel != "" {
		// Condition node: match label ("true"/"false") to edge condition field
		for _, edge := range outEdges {
			if edge.Condition == output.NextEdgeLabel {
				return edge.TargetNodeKey, nil
			}
		}
		// Fallback: unconditional edge
		for _, edge := range outEdges {
			if edge.Condition == "" {
				return edge.TargetNodeKey, nil
			}
		}
		return "", fmt.Errorf("no edge matched label %q", output.NextEdgeLabel)
	}

	// Non-condition node: take first edge
	return outEdges[0].TargetNodeKey, nil
}

// failRun marks the run as FAILED and returns nil (the error was the run failure itself).
func (e *Engine) failRun(ctx context.Context, runID, code, message, nodeKey string) error {
	runErr := &wf.RunError{Code: code, Message: message, NodeKey: nodeKey}
	return e.repo.UpdateRunStatus(ctx, runID, wf.RunStatusFailed, runErr)
}
