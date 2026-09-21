package engine

import (
	"fmt"

	wf "verification-platform/internal/domain/workflow"
)

// ValidationError represents a single graph validation failure.
type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (v ValidationError) Error() string { return fmt.Sprintf("%s: %s", v.Code, v.Message) }

// ValidationResult is the output of the graph validator.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors"`
	Warnings []ValidationError `json:"warnings"`
}

// Validate checks a workflow graph for correctness before publication.
func Validate(nodes []*wf.WorkflowNode, edges []*wf.WorkflowEdge) ValidationResult {
	var errs []ValidationError

	// Build lookup maps
	nodeByKey := map[string]*wf.WorkflowNode{}
	for _, n := range nodes {
		if _, dup := nodeByKey[n.NodeKey]; dup {
			errs = append(errs, ValidationError{
				Code:    "DUPLICATE_NODE_KEY",
				Message: fmt.Sprintf("duplicate node key: %s", n.NodeKey),
			})
		}
		nodeByKey[n.NodeKey] = n
	}

	// Check all node types are registered
	for _, n := range nodes {
		if !wf.AllNodeTypes[n.Type] {
			errs = append(errs, ValidationError{
				Code:    "INVALID_NODE_TYPE",
				Message: fmt.Sprintf("node %s has unknown type: %s", n.NodeKey, n.Type),
			})
		}
	}

	// Exactly one START node
	startCount := 0
	var startKey string
	for _, n := range nodes {
		if n.Type == wf.NodeTypeStart {
			startCount++
			startKey = n.NodeKey
		}
	}
	if startCount == 0 {
		errs = append(errs, ValidationError{Code: "NO_START_NODE", Message: "workflow must have exactly one START node"})
	} else if startCount > 1 {
		errs = append(errs, ValidationError{Code: "MULTIPLE_START_NODES", Message: "workflow must have exactly one START node"})
	}

	// At least one END node
	endCount := 0
	for _, n := range nodes {
		if n.Type == wf.NodeTypeEnd {
			endCount++
		}
	}
	if endCount == 0 {
		errs = append(errs, ValidationError{Code: "NO_END_NODE", Message: "workflow must have at least one END node"})
	}

	// Edge validation: sources and targets must exist
	edgeIDs := map[string]bool{}
	for _, e := range edges {
		eid := e.ID.String()
		if edgeIDs[eid] {
			errs = append(errs, ValidationError{Code: "DUPLICATE_EDGE", Message: fmt.Sprintf("duplicate edge id: %s", eid)})
		}
		edgeIDs[eid] = true

		if _, ok := nodeByKey[e.SourceNodeKey]; !ok {
			errs = append(errs, ValidationError{
				Code:    "MISSING_SOURCE_NODE",
				Message: fmt.Sprintf("edge source node not found: %s", e.SourceNodeKey),
			})
		}
		if _, ok := nodeByKey[e.TargetNodeKey]; !ok {
			errs = append(errs, ValidationError{
				Code:    "MISSING_TARGET_NODE",
				Message: fmt.Sprintf("edge target node not found: %s", e.TargetNodeKey),
			})
		}
	}

	// Build adjacency list (source → targets)
	adj := map[string][]string{}
	for _, e := range edges {
		adj[e.SourceNodeKey] = append(adj[e.SourceNodeKey], e.TargetNodeKey)
	}

	// Orphan detection: every non-START node must be reachable from START
	if startCount == 1 {
		reachable := map[string]bool{}
		var bfs func(key string)
		bfs = func(key string) {
			if reachable[key] {
				return
			}
			reachable[key] = true
			for _, next := range adj[key] {
				bfs(next)
			}
		}
		bfs(startKey)

		for _, n := range nodes {
			if !reachable[n.NodeKey] {
				errs = append(errs, ValidationError{
					Code:    "ORPHAN_NODE",
					Message: fmt.Sprintf("node %s is not reachable from START", n.NodeKey),
				})
			}
		}
	}

	// Cycle detection (DFS)
	if startCount == 1 {
		visited := map[string]bool{}
		inStack := map[string]bool{}
		var hasCycle bool
		var dfs func(key string)
		dfs = func(key string) {
			if hasCycle {
				return
			}
			visited[key] = true
			inStack[key] = true
			for _, next := range adj[key] {
				if !visited[next] {
					dfs(next)
				} else if inStack[next] {
					hasCycle = true
					return
				}
			}
			inStack[key] = false
		}
		dfs(startKey)
		if hasCycle {
			errs = append(errs, ValidationError{Code: "CYCLE_DETECTED", Message: "workflow graph contains a cycle"})
		}
	}

	return ValidationResult{Valid: len(errs) == 0, Errors: errs}
}
