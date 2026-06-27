package node

import (
	"encoding/json"
	"fmt"

	"confluence-mcp-v10.2.14/internal/mcpaggregator/pipeline"
)

// ReturnValue resolves the return source and returns the raw value.
func ReturnValue(step *pipeline.StepConfig, rctx pipeline.StepContext) (interface{}, error) {
	cfg := step.Return
	return rctx.ResolvePath(cfg.Source)
}

// ReturnNode converts a value to a CallToolResult.
func ReturnNode(val interface{}) (*pipeline.CallToolResult, error) {
	switch v := val.(type) {
	case string:
		return &pipeline.CallToolResult{
			Content: []pipeline.ContentItem{{Type: "text", Text: v}},
		}, nil
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return &pipeline.CallToolResult{
				Content: []pipeline.ContentItem{{Type: "text", Text: fmt.Sprintf("%v", v)}},
			}, nil
		}
		return &pipeline.CallToolResult{
			Content: []pipeline.ContentItem{{Type: "text", Text: string(b)}},
		}, nil
	}
}
