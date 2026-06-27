package runtime

import (
	"context"
	"fmt"

	"jira-mcp-v10.7.4/internal/mcpaggregator/node"
	"jira-mcp-v10.7.4/internal/mcpaggregator/pipeline"
)

// Executor runs a pipeline of steps.
type Executor struct {
	registry pipeline.ToolRegistry
}

// NewExecutor creates a pipeline executor.
func NewExecutor(registry pipeline.ToolRegistry) *Executor {
	return &Executor{registry: registry}
}

// Execute runs a pipeline and returns the final result.
func (e *Executor) Execute(ctx context.Context, steps []pipeline.StepConfig, input map[string]interface{}) (*pipeline.CallToolResult, error) {
	rctx := NewContext(input)

	for _, step := range steps {
		result, err := e.ExecuteStep(ctx, &step, rctx)
		if err != nil {
			return nil, fmt.Errorf("step %q: %w", step.Name, err)
		}

		if step.Output != "" {
			rctx.SetOutput(step.Output, result)
		}
		rctx.SetOutput(step.Name, result)

		if step.Type == "return" {
			return node.ReturnNode(result)
		}
	}

	return nil, fmt.Errorf("pipeline completed without a return step")
}

// ExecuteStep executes a single pipeline step (implements pipeline.StepExecutor).
func (e *Executor) ExecuteStep(ctx context.Context, step *pipeline.StepConfig, rctx pipeline.StepContext) (interface{}, error) {
	switch step.Type {
	case "call":
		return node.CallNode(ctx, step, rctx, e.registry)
	case "map":
		return node.MapNode(ctx, step, rctx, e)
	case "transform":
		return node.TransformNode(step, rctx)
	case "merge":
		return node.MergeNode(step, rctx)
	case "return":
		return node.ReturnValue(step, rctx)
	default:
		return nil, fmt.Errorf("unknown step type %q", step.Type)
	}
}
