package node

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"

	"confluence-mcp-v10.2.14/internal/mcpaggregator/pipeline"
)

// JQNode evaluates a jq expression against pipeline data.
// It resolves `from` as the input (`.`), `vars` as `$variable` bindings,
// compiles the `expr`, and returns the result.
func JQNode(step *pipeline.StepConfig, rctx pipeline.StepContext) (interface{}, error) {
	spec := step.Spec

	// Resolve the input data (the `.` in jq)
	var input interface{}
	if spec.From != "" {
		resolved, err := resolveFrom(spec.From, rctx)
		if err != nil {
			return nil, fmt.Errorf("jq from: %w", err)
		}
		input = resolved
	}

	// Parse the jq expression
	query, err := gojq.Parse(spec.Expr)
	if err != nil {
		return nil, fmt.Errorf("jq parse error: %w", err)
	}

	// Resolve variables and build gojq bindings
	varNames, varValues, err := resolveJQVars(spec.Vars, rctx)
	if err != nil {
		return nil, fmt.Errorf("jq vars: %w", err)
	}

	// Run the query
	code, err := gojq.Compile(query, gojq.WithVariables(varNames))
	if err != nil {
		return nil, fmt.Errorf("jq compile error: %w", err)
	}

	iter := code.Run(input, varValues...)

	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("jq execution error: %w", err)
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

// resolveFrom resolves the From field of a jq/return/emit step.
// It handles $ref (step outputs, input, item) and literal JSON strings.
func resolveFrom(from string, rctx pipeline.StepContext) (interface{}, error) {
	from = strings.TrimSpace(from)
	if from == "" {
		return nil, nil
	}

	// Try $ref resolution first
	if strings.HasPrefix(from, "$") {
		return rctx.Resolve(from)
	}

	// Try parsing as JSON literal
	if (strings.HasPrefix(from, "{") && strings.HasSuffix(from, "}")) ||
		(strings.HasPrefix(from, "[") && strings.HasSuffix(from, "]")) {
		var v interface{}
		if err := json.Unmarshal([]byte(from), &v); err == nil {
			return v, nil
		}
	}

	// Plain string literal
	return from, nil
}

// resolveJQVars resolves variable bindings for a jq expression.
// Returns variable names and their values for gojq.WithVariables / code.Run.
func resolveJQVars(vars map[string]interface{}, rctx pipeline.StepContext) ([]string, []interface{}, error) {
	if len(vars) == 0 {
		return nil, nil, nil
	}

	varNames := make([]string, 0, len(vars))
	varValues := make([]interface{}, 0, len(vars))

	for name, raw := range vars {
		var val interface{}
		switch v := raw.(type) {
		case string:
			resolved, err := rctx.Resolve(v)
			if err != nil {
				return nil, nil, fmt.Errorf("var %q: %w", name, err)
			}
			val = resolved
		default:
			val = v
		}
		varNames = append(varNames, "$"+name)
		varValues = append(varValues, val)
	}

	return varNames, varValues, nil
}
