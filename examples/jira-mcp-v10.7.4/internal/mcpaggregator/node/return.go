package node

import (
	"encoding/json"
	"fmt"

	"github.com/itchyny/gojq"

	"jira-mcp-v10.7.4/internal/mcpaggregator/pipeline"
)

// ReturnValue resolves the return step's value. If an `expr` is present,
// it evaluates the jq expression against the resolved `from` data with `vars`.
// Otherwise, it resolves `from` and returns the value as-is.
func ReturnValue(step *pipeline.StepConfig, rctx pipeline.StepContext) (interface{}, error) {
	spec := step.Spec

	if spec.Expr == "" {
		// Simple return: resolve `from` and return as-is
		return resolveFrom(spec.From, rctx)
	}

	// JQ return: resolve from, vars, evaluate expr
	var input interface{}
	if spec.From != "" {
		resolved, err := resolveFrom(spec.From, rctx)
		if err != nil {
			return nil, fmt.Errorf("return from: %w", err)
		}
		input = resolved
	}

	query, err := gojq.Parse(spec.Expr)
	if err != nil {
		return nil, fmt.Errorf("return jq parse error: %w", err)
	}

	varNames, varValues, err := resolveJQVars(spec.Vars, rctx)
	if err != nil {
		return nil, fmt.Errorf("return vars: %w", err)
	}

	code, err := gojq.Compile(query, gojq.WithVariables(varNames))
	if err != nil {
		return nil, fmt.Errorf("return jq compile error: %w", err)
	}

	iter := code.Run(input, varValues...)

	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("return jq execution error: %w", err)
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
