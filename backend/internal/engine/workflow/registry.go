package engine

import (
	"context"
)

// ExecutionInput is the data provided to a node at execution time.
type ExecutionInput struct {
	NodeKey    string                            `json:"node_key"`
	RunID      string                            `json:"run_id"`
	NodeConfig map[string]interface{}            `json:"node_config"`
	Context    map[string]interface{}            `json:"context"`    // verification + workflow metadata
	NodeOutputs map[string]map[string]interface{} `json:"node_outputs"` // outputs of prior nodes keyed by node_key
}

// ExecutionOutput is what a node returns after running.
type ExecutionOutput struct {
	Data          map[string]interface{} `json:"data"`
	NextEdgeLabel string                 `json:"next_edge_label,omitempty"` // used by CONDITION to select branch
	WaitForEvent  bool                   `json:"wait_for_event,omitempty"`  // used by future VOICE_AGENT etc.
}

// NodeExecutor is the interface every node type must implement.
type NodeExecutor interface {
	Type() string
	Execute(ctx context.Context, input ExecutionInput) (ExecutionOutput, error)
}

// Registry holds all registered node executors.
type Registry struct {
	executors map[string]NodeExecutor
}

func NewRegistry() *Registry {
	return &Registry{executors: map[string]NodeExecutor{}}
}

// Register adds an executor. Returns an error if the type is already registered.
func (r *Registry) Register(e NodeExecutor) {
	r.executors[e.Type()] = e
}

// Get returns the executor for the given node type.
func (r *Registry) Get(nodeType string) (NodeExecutor, bool) {
	e, ok := r.executors[nodeType]
	return e, ok
}
