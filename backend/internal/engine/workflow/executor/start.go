package executor

import (
	"context"
	engine "verification-platform/internal/engine/workflow"
)

// StartExecutor is the entry point for every workflow run.
type StartExecutor struct{}

func (s *StartExecutor) Type() string { return "START" }

func (s *StartExecutor) Execute(_ context.Context, input engine.ExecutionInput) (engine.ExecutionOutput, error) {
	// Pass the full context through so downstream nodes have access to it.
	return engine.ExecutionOutput{Data: input.Context}, nil
}
