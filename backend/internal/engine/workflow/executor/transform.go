package executor

import (
	"context"
	"fmt"

	engine "verification-platform/internal/engine/workflow"
)

// TransformExecutor remaps keys from the execution context according to
// a "mappings" config of the form {"output_key": "input.path"}.
// It only supports simple dot-notation resolution over the execution context.
type TransformExecutor struct{}

func (t *TransformExecutor) Type() string { return "TRANSFORM" }

func (t *TransformExecutor) Execute(_ context.Context, input engine.ExecutionInput) (engine.ExecutionOutput, error) {
	mappings, _ := input.NodeConfig["mappings"].(map[string]interface{})
	result := map[string]interface{}{}

	for outputKey, pathVal := range mappings {
		path, ok := pathVal.(string)
		if !ok {
			continue
		}
		val := resolvePath(input.Context, input.NodeOutputs, path)
		result[outputKey] = val
	}

	return engine.ExecutionOutput{Data: result}, nil
}

// resolvePath resolves a dot-separated path from context or prior node outputs.
// Supports two forms:
//   "key"          → input.Context["key"]
//   "node.output.key" → input.NodeOutputs["node"]["output"]["key"]
func resolvePath(ctx map[string]interface{}, nodeOutputs map[string]map[string]interface{}, path string) interface{} {
	parts := splitDot(path)
	if len(parts) == 0 {
		return nil
	}

	// Check node outputs first
	if nodeOut, ok := nodeOutputs[parts[0]]; ok && len(parts) > 1 {
		return deepGet(nodeOut, parts[1:])
	}

	return deepGet(ctx, parts)
}

func deepGet(m map[string]interface{}, keys []string) interface{} {
	if len(keys) == 0 {
		return nil
	}
	val, ok := m[keys[0]]
	if !ok {
		return nil
	}
	if len(keys) == 1 {
		return val
	}
	nested, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}
	return deepGet(nested, keys[1:])
}

func splitDot(s string) []string {
	var parts []string
	cur := ""
	for _, c := range s {
		if c == '.' {
			if cur != "" {
				parts = append(parts, cur)
				cur = ""
			}
		} else {
			cur += fmt.Sprintf("%c", c)
		}
	}
	if cur != "" {
		parts = append(parts, cur)
	}
	return parts
}
