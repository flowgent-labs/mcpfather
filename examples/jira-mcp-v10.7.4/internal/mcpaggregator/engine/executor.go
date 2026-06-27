package engine

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
			return nil, fmt.Errorf("Step %q: %w", step.ID, err)
		}

		// Check require constraints
		if err := checkRequire(&step, result); err != nil {
			return nil, fmt.Errorf("Step %q validation failed: %w", step.ID, err)
		}

		// Store output under step ID for downstream references
		rctx.SetOutput(step.ID, result)

		if step.Kind == "return" {
			return node.ReturnNode(result)
		}
	}

	return nil, fmt.Errorf("Pipeline completed without a return step")
}

// ExecuteStep executes a single pipeline step (implements pipeline.StepExecutor).
func (e *Executor) ExecuteStep(ctx context.Context, step *pipeline.StepConfig, rctx pipeline.StepContext) (interface{}, error) {
	switch step.Kind {
	case "call":
		return node.CallNode(ctx, step, rctx, e.registry)
	case "jq":
		return node.JQNode(step, rctx)
	case "foreach":
		return node.ForeachNode(ctx, step, rctx, e)
	case "return":
		return node.ReturnValue(step, rctx)
	case "emit":
		return node.EmitNode(step, rctx)
	default:
		return nil, fmt.Errorf("Unknown step kind %q", step.Kind)
	}
}

// checkRequire validates a step's result against its require constraints.
func checkRequire(step *pipeline.StepConfig, result interface{}) error {
	if step.Require == nil {
		return nil
	}
	if step.Require.NonEmpty {
		if isEmpty(result) {
			if step.Require.Message != "" {
				return fmt.Errorf("%s", step.Require.Message)
			}
			return fmt.Errorf("Step %q: result must not be empty", step.ID)
		}
	}
	return nil
}

// isEmpty checks whether a value is semantically empty.
func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case []interface{}:
		return len(val) == 0
	case map[string]interface{}:
		return len(val) == 0
	case bool:
		return !val
	case int, int64, float64:
		return false
	default:
		return false
	}
}
