package node

import (
	"fmt"

	"github.com/itchyny/gojq"
	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/pipeline"
)

// EmitNode evaluates a jq expression and returns the result — used within
// foreach sub-pipelines to produce transformed output elements.
func EmitNode(step *pipeline.StepConfig, rctx pipeline.StepContext) (interface{}, error) {
	spec := step.Spec

	if spec.Expr == "" {
		// No expression: resolve `from` and return as-is
		return resolveFrom(spec.From, rctx)
	}

	// Resolve the input data
	var input interface{}
	if spec.From != "" {
		resolved, err := resolveFrom(spec.From, rctx)
		if err != nil {
			return nil, fmt.Errorf("emit from: %w", err)
		}
		input = resolved
	}

	// Parse and compile jq expression
	query, err := gojq.Parse(spec.Expr)
	if err != nil {
		return nil, fmt.Errorf("emit jq parse error: %w", err)
	}

	varNames, varValues, err := resolveJQVars(spec.Vars, rctx)
	if err != nil {
		return nil, fmt.Errorf("emit vars: %w", err)
	}

	code, err := gojq.Compile(query, gojq.WithVariables(varNames))
	if err != nil {
		return nil, fmt.Errorf("emit jq compile error: %w", err)
	}

	iter := code.Run(input, varValues...)

	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("emit jq execution error: %w", err)
		}
		results = append(results, v)
	}

	if len(results) == 0 {
		return nil, nil
	}
	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}
