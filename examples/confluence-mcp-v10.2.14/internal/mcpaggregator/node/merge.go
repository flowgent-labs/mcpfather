package node

import (
	"fmt"
	"strings"

	"confluence-mcp-v10.2.14/internal/mcpaggregator/pipeline"
)

// MergeNode merges data from one path into another within the runtime context.
func MergeNode(step *pipeline.StepConfig, rctx pipeline.StepContext) (interface{}, error) {
	cfg := step.Merge

	fromVal, err := rctx.ResolvePath(cfg.From)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve merge.from: %w", err)
	}

	parts := splitPath(cfg.To)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid merge.to path: %q", cfg.To)
	}

	// Resolve the target root (everything except the last segment).
	// "echoResult.output.greeting_data" → resolve "echoResult.output" to get the map.
	rootPath := strings.Join(parts[:len(parts)-1], ".")
	targetObj, err := rctx.ResolvePath(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve merge.to root: %w", err)
	}

	targetMap, ok := targetObj.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("merge.to root must resolve to an object, got %T", targetObj)
	}

	key := parts[len(parts)-1]
	targetMap[key] = fromVal
	return targetMap, nil
}

// ApplyMerge writes the merged value into the target object at the given path.
// Only the last segment of toPath is used as the key.
func ApplyMerge(target map[string]interface{}, toPath string, value interface{}) {
	parts := splitPath(toPath)
	if len(parts) == 0 {
		return
	}
	key := parts[len(parts)-1]
	target[key] = value
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	var parts []string
	current := ""
	for _, c := range path {
		if c == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
