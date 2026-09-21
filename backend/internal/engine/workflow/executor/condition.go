package executor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	engine "verification-platform/internal/engine/workflow"
)

// ConditionExecutor evaluates a safe boolean expression and sets NextEdgeLabel
// to "true" or "false" so the engine can select the correct outgoing edge.
//
// Supported expression syntax (whitespace-insensitive):
//   <path> == "<value>"
//   <path> != "<value>"
//   <path> >= <number>
//   <path> <= <number>
//   <path> > <number>
//   <path> < <number>
//
// <path> uses the same dot-notation as TransformExecutor.
// Compound expressions are NOT supported in Phase 04 to keep it safe.
type ConditionExecutor struct{}

func (c *ConditionExecutor) Type() string { return "CONDITION" }

func (c *ConditionExecutor) Execute(_ context.Context, input engine.ExecutionInput) (engine.ExecutionOutput, error) {
	expr, _ := input.NodeConfig["expression"].(string)
	result := evalExpression(expr, input.Context, input.NodeOutputs)
	label := "false"
	if result {
		label = "true"
	}
	return engine.ExecutionOutput{
		Data:          map[string]interface{}{"result": result},
		NextEdgeLabel: label,
	}, nil
}

// evalExpression evaluates a single safe binary expression.
func evalExpression(expr string, ctx map[string]interface{}, nodeOutputs map[string]map[string]interface{}) bool {
	expr = strings.TrimSpace(expr)

	for _, op := range []string{"==", "!=", ">=", "<=", ">", "<"} {
		idx := strings.Index(expr, op)
		if idx < 0 {
			continue
		}
		lhs := strings.TrimSpace(expr[:idx])
		rhs := strings.TrimSpace(expr[idx+len(op):])

		lval := resolvePath(ctx, nodeOutputs, lhs)
		lStr := fmt.Sprintf("%v", lval)

		// String comparison (rhs quoted)
		if strings.HasPrefix(rhs, `"`) && strings.HasSuffix(rhs, `"`) {
			rStr := rhs[1 : len(rhs)-1]
			switch op {
			case "==":
				return lStr == rStr
			case "!=":
				return lStr != rStr
			}
			return false
		}

		// Numeric comparison
		lNum, lErr := strconv.ParseFloat(lStr, 64)
		rNum, rErr := strconv.ParseFloat(rhs, 64)
		if lErr != nil || rErr != nil {
			return false
		}
		switch op {
		case "==":
			return lNum == rNum
		case "!=":
			return lNum != rNum
		case ">=":
			return lNum >= rNum
		case "<=":
			return lNum <= rNum
		case ">":
			return lNum > rNum
		case "<":
			return lNum < rNum
		}
	}
	return false
}
