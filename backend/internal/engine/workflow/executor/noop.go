package executor

import (
	"context"
	engine "verification-platform/internal/engine/workflow"
)

// NoopExecutor passes its input data through unchanged.
type NoopExecutor struct{}

func (n *NoopExecutor) Type() string { return "NOOP" }

func (n *NoopExecutor) Execute(_ context.Context, input engine.ExecutionInput) (engine.ExecutionOutput, error) {
	data := map[string]interface{}{"noop": true}
	for k, v := range input.Context {
		data[k] = v
	}
	return engine.ExecutionOutput{Data: data}, nil
}
