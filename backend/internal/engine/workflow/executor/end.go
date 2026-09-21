package executor

import (
	"context"
	engine "verification-platform/internal/engine/workflow"
)

// EndExecutor marks the successful completion of a workflow path.
type EndExecutor struct{}

func (e *EndExecutor) Type() string { return "END" }

func (e *EndExecutor) Execute(_ context.Context, input engine.ExecutionInput) (engine.ExecutionOutput, error) {
	return engine.ExecutionOutput{Data: map[string]interface{}{"status": "completed"}}, nil
}
