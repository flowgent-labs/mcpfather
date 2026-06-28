package pipeline

import (
	"fmt"
	"strings"
)

// Validate checks a list of step configs for structural correctness.
func Validate(steps []StepConfig) error {
	ids := make(map[string]bool)

	for i, step := range steps {
		if step.ID == "" {
			return fmt.Errorf("Step %d: id is required", i)
		}
		if ids[step.ID] {
			return fmt.Errorf("Step %q: duplicate id", step.ID)
		}
		ids[step.ID] = true

		if err := validateStep(step); err != nil {
			return fmt.Errorf("Step %q: %w", step.ID, err)
		}
	}
	return nil
}

func validateStep(step StepConfig) error {
	switch step.Kind {
	case "call":
		if step.Spec.Tool == "" {
			return fmt.Errorf("Kind 'call' requires spec.tool")
		}
		if step.Spec.Args == nil {
			return fmt.Errorf("Kind 'call' requires spec.args")
		}
	case "jq":
		if step.Spec.Expr == "" {
			return fmt.Errorf("Kind 'jq' requires spec.expr")
		}
	case "foreach":
		if step.Spec.In == "" {
			return fmt.Errorf("Kind 'foreach' requires spec.in")
		}
		if step.Spec.As == "" {
			return fmt.Errorf("Kind 'foreach' requires spec.as")
		}
		if len(step.Spec.Pipeline) == 0 {
			return fmt.Errorf("Kind 'foreach' requires spec.pipeline with at least one step")
		}
		if err := validateForeachPipeline(step.Spec.Pipeline); err != nil {
			return fmt.Errorf("foreach.pipeline: %w", err)
		}
	case "return":
		// from is optional (may use only vars + expr, or literal)
	case "emit":
		// from is optional (may use item directly)
	default:
		return fmt.Errorf("Unknown step kind %q (valid: call, jq, foreach, return, emit)", step.Kind)
	}
	return nil
}

// validateForeachPipeline validates a foreach sub-pipeline.
// It must contain at least one emit step and must end with emit.
func validateForeachPipeline(steps []StepConfig) error {
	ids := make(map[string]bool)
	hasEmit := false

	for i, s := range steps {
		if s.ID == "" {
			return fmt.Errorf("Step %d: id is required", i)
		}
		if ids[s.ID] {
			return fmt.Errorf("Step %q: duplicate id", s.ID)
		}
		ids[s.ID] = true

		if s.Kind == "emit" {
			hasEmit = true
			continue
		}
		if s.Kind == "return" {
			return fmt.Errorf("Step %q: 'return' is not allowed inside foreach; use 'emit' instead", s.ID)
		}
		if s.Kind == "foreach" {
			return fmt.Errorf("Step %q: nested foreach is not supported", s.ID)
		}
		// Recurse to validate other step kinds
		if err := validateStep(s); err != nil {
			return fmt.Errorf("Step %q: %w", s.ID, err)
		}
	}

	if !hasEmit {
		return fmt.Errorf("Foreach pipeline must contain at least one 'emit' step")
	}
	return nil
}

// ValidateReferences checks that all $ references point to known sources.
// Known sources: input, step IDs, foreach as names.
func ValidateReferences(steps []StepConfig) error {
	return validateRefsInPipeline(steps, nil)
}

func validateRefsInPipeline(steps []StepConfig, parentIDs map[string]bool) error {
	known := map[string]bool{
		"input": true, // always available
	}
	for _, step := range steps {
		known[step.ID] = true
	}
	if parentIDs != nil {
		for id := range parentIDs {
			known[id] = true
		}
	}

	for _, step := range steps {
		refs := extractRefs(step)
		for _, ref := range refs {
			root := strings.SplitN(ref, ".", 2)[0]
			if !known[root] {
				return fmt.Errorf("Step %q: unresolved reference %q (root %q is not input, a step id, or a foreach variable)", step.ID, ref, root)
			}
		}

		// Validate foreach sub-pipelines
		if step.Kind == "foreach" {
			// Add the foreach 'as' name to known refs visible inside the sub-pipeline
			foreachKnown := make(map[string]bool)
			for id := range known {
				foreachKnown[id] = true
			}
			if step.Spec.As != "" {
				foreachKnown[step.Spec.As] = true
			}
			if err := validateRefsInPipeline(step.Spec.Pipeline, foreachKnown); err != nil {
				return fmt.Errorf("Step %q foreach.pipeline: %w", step.ID, err)
			}
		}
	}
	return nil
}

func extractRefs(step StepConfig) []string {
	var refs []string

	// Extract from args
	for _, v := range step.Spec.Args {
		refs = append(refs, extractStringRefs(v)...)
	}

	// Extract from vars
	for _, v := range step.Spec.Vars {
		refs = append(refs, extractStringRefs(v)...)
	}

	// Extract from from, in
	refs = append(refs, extractDollarRefs(step.Spec.From)...)
	refs = append(refs, extractDollarRefs(step.Spec.In)...)

	// Extract from concurrency if it's a string ref
	if s, ok := step.Spec.Concurrency.(string); ok {
		refs = append(refs, extractDollarRefs(s)...)
	}

	return refs
}

func extractDollarRefs(s string) []string {
	if !strings.HasPrefix(s, "$") {
		return nil
	}
	ref := strings.TrimPrefix(s, "$")
	if ref == "" {
		return nil
	}
	return []string{ref}
}

func extractStringRefs(v interface{}) []string {
	switch val := v.(type) {
	case string:
		if strings.HasPrefix(val, "$") {
			return []string{strings.TrimPrefix(val, "$")}
		}
	case map[string]interface{}:
		var refs []string
		for _, sv := range val {
			refs = append(refs, extractStringRefs(sv)...)
		}
		return refs
	case []interface{}:
		var refs []string
		for _, item := range val {
			refs = append(refs, extractStringRefs(item)...)
		}
		return refs
	}
	return nil
}
