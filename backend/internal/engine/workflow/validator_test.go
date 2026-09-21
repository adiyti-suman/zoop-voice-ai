package engine

import (
	"testing"
	"time"

	"github.com/google/uuid"
	wf "verification-platform/internal/domain/workflow"
)

func makeNode(key string, t wf.NodeType) *wf.WorkflowNode {
	return &wf.WorkflowNode{ID: uuid.New(), NodeKey: key, Type: t, Name: key, CreatedAt: time.Now()}
}

func makeEdge(src, tgt string) *wf.WorkflowEdge {
	return &wf.WorkflowEdge{ID: uuid.New(), SourceNodeKey: src, TargetNodeKey: tgt, CreatedAt: time.Now()}
}

// ── Failure cases ─────────────────────────────────────────────────────────────

func TestValidate_NoStartNode(t *testing.T) {
	nodes := []*wf.WorkflowNode{makeNode("end", wf.NodeTypeEnd)}
	r := Validate(nodes, nil)
	assertError(t, r, "NO_START_NODE")
}

func TestValidate_MultipleStartNodes(t *testing.T) {
	nodes := []*wf.WorkflowNode{
		makeNode("s1", wf.NodeTypeStart), makeNode("s2", wf.NodeTypeStart), makeNode("e", wf.NodeTypeEnd),
	}
	r := Validate(nodes, nil)
	assertError(t, r, "MULTIPLE_START_NODES")
}

func TestValidate_NoEndNode(t *testing.T) {
	nodes := []*wf.WorkflowNode{makeNode("s", wf.NodeTypeStart)}
	r := Validate(nodes, nil)
	assertError(t, r, "NO_END_NODE")
}

func TestValidate_MissingSourceNode(t *testing.T) {
	nodes := []*wf.WorkflowNode{makeNode("s", wf.NodeTypeStart), makeNode("e", wf.NodeTypeEnd)}
	edges := []*wf.WorkflowEdge{makeEdge("ghost", "e")}
	r := Validate(nodes, edges)
	assertError(t, r, "MISSING_SOURCE_NODE")
}

func TestValidate_MissingTargetNode(t *testing.T) {
	nodes := []*wf.WorkflowNode{makeNode("s", wf.NodeTypeStart), makeNode("e", wf.NodeTypeEnd)}
	edges := []*wf.WorkflowEdge{makeEdge("s", "ghost")}
	r := Validate(nodes, edges)
	assertError(t, r, "MISSING_TARGET_NODE")
}

func TestValidate_DuplicateNodeKey(t *testing.T) {
	nodes := []*wf.WorkflowNode{makeNode("s", wf.NodeTypeStart), makeNode("s", wf.NodeTypeEnd)}
	r := Validate(nodes, nil)
	assertError(t, r, "DUPLICATE_NODE_KEY")
}

func TestValidate_InvalidNodeType(t *testing.T) {
	nodes := []*wf.WorkflowNode{
		makeNode("s", wf.NodeTypeStart),
		makeNode("e", wf.NodeTypeEnd),
		{ID: uuid.New(), NodeKey: "bad", Type: "TOTALLY_FAKE", CreatedAt: time.Now()},
	}
	edges := []*wf.WorkflowEdge{makeEdge("s", "bad"), makeEdge("bad", "e")}
	r := Validate(nodes, edges)
	assertError(t, r, "INVALID_NODE_TYPE")
}

func TestValidate_OrphanNode(t *testing.T) {
	nodes := []*wf.WorkflowNode{
		makeNode("s", wf.NodeTypeStart),
		makeNode("e", wf.NodeTypeEnd),
		makeNode("orphan", wf.NodeTypeNoop),
	}
	edges := []*wf.WorkflowEdge{makeEdge("s", "e")}
	r := Validate(nodes, edges)
	assertError(t, r, "ORPHAN_NODE")
}

func TestValidate_CycleDetected(t *testing.T) {
	nodes := []*wf.WorkflowNode{
		makeNode("s", wf.NodeTypeStart),
		makeNode("a", wf.NodeTypeNoop),
		makeNode("b", wf.NodeTypeNoop),
		makeNode("e", wf.NodeTypeEnd),
	}
	edges := []*wf.WorkflowEdge{
		makeEdge("s", "a"), makeEdge("a", "b"), makeEdge("b", "a"), makeEdge("b", "e"),
	}
	r := Validate(nodes, edges)
	assertError(t, r, "CYCLE_DETECTED")
}

// ── Pass cases ────────────────────────────────────────────────────────────────

func TestValidate_StartEnd(t *testing.T) {
	nodes := []*wf.WorkflowNode{makeNode("s", wf.NodeTypeStart), makeNode("e", wf.NodeTypeEnd)}
	edges := []*wf.WorkflowEdge{makeEdge("s", "e")}
	r := Validate(nodes, edges)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %+v", r.Errors)
	}
}

func TestValidate_StartNoopEnd(t *testing.T) {
	nodes := []*wf.WorkflowNode{makeNode("s", wf.NodeTypeStart), makeNode("n", wf.NodeTypeNoop), makeNode("e", wf.NodeTypeEnd)}
	edges := []*wf.WorkflowEdge{makeEdge("s", "n"), makeEdge("n", "e")}
	r := Validate(nodes, edges)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %+v", r.Errors)
	}
}

func TestValidate_StartTransformEnd(t *testing.T) {
	nodes := []*wf.WorkflowNode{makeNode("s", wf.NodeTypeStart), makeNode("t", wf.NodeTypeTransform), makeNode("e", wf.NodeTypeEnd)}
	edges := []*wf.WorkflowEdge{makeEdge("s", "t"), makeEdge("t", "e")}
	r := Validate(nodes, edges)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %+v", r.Errors)
	}
}

func TestValidate_BranchingCondition(t *testing.T) {
	nodes := []*wf.WorkflowNode{
		makeNode("s", wf.NodeTypeStart), makeNode("c", wf.NodeTypeCondition),
		makeNode("e1", wf.NodeTypeEnd), makeNode("e2", wf.NodeTypeEnd),
	}
	edges := []*wf.WorkflowEdge{makeEdge("s", "c"), makeEdge("c", "e1"), makeEdge("c", "e2")}
	r := Validate(nodes, edges)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %+v", r.Errors)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func assertError(t *testing.T, r ValidationResult, code string) {
	t.Helper()
	if r.Valid {
		t.Errorf("expected validation failure with code %s, but got valid=true", code)
		return
	}
	for _, e := range r.Errors {
		if e.Code == code {
			return
		}
	}
	t.Errorf("expected error code %s in %+v", code, r.Errors)
}
