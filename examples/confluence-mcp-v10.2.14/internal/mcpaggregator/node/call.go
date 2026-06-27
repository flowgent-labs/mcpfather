package node

import (
	"context"
	"encoding/json"
	"fmt"

	"confluence-mcp-v10.2.14/internal/mcpaggregator/pipeline"
)

// CallNode executes a call step — invoking a native MCP tool.
func CallNode(ctx context.Context, step *pipeline.StepConfig, rctx pipeline.StepContext, registry pipeline.ToolRegistry) (interface{}, error) {
	cfg := step.Call

	resolvedArgs, err := rctx.ResolveMap(cfg.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve call args: %w", err)
	}

	result, err := registry.CallTool(ctx, cfg.Tool, resolvedArgs)
	if err != nil {
		return nil, fmt.Errorf("tool %q failed: %w", cfg.Tool, err)
	}

	if result.IsError {
		return nil, fmt.Errorf("tool %q returned error", cfg.Tool)
	}

	text := extractTextContent(result)
	if text == "" {
		return result, nil
	}

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
