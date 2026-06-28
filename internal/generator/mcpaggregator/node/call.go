package node

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/pipeline"
)

// CallNode executes a call step — invoking a native MCP tool.
func CallNode(ctx context.Context, step *pipeline.StepConfig, rctx pipeline.StepContext, registry pipeline.ToolRegistry) (interface{}, error) {
	spec := step.Spec

	resolvedArgs, err := rctx.ResolveMap(spec.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve call args: %w", err)
	}

	result, err := registry.CallTool(ctx, spec.Tool, resolvedArgs)
	if err != nil {
		return nil, fmt.Errorf("tool %q failed: %w", spec.Tool, err)
	}

	if result.IsError {
		return nil, fmt.Errorf("tool %q returned error", spec.Tool)
	}

	text := extractTextContent(result)
	if text == "" {
		return result, nil
	}

	// Parse JSON if spec.parse is "json" or the result looks like JSON
	if spec.Parse == "json" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(text), &parsed); err != nil {
			return nil, fmt.Errorf("failed to parse tool response as JSON: %w", err)
		}
		return parsed, nil
	}

	// Auto-detect JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(text), &parsed); err == nil {
		return parsed, nil
	}
	return text, nil
}

func extractTextContent(result *pipeline.CallToolResult) string {
	for _, c := range result.Content {
		if c.Type == "text" {
			return c.Text
		}
	}
	return ""
}
