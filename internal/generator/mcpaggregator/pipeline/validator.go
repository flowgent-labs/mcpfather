package pipeline

import (
	"fmt"
	"regexp"
	"strings"
)

var refPattern = regexp.MustCompile(`\{\{\s*([^}]+)\s*\}\}`)

// Validate checks a list of step configs for structural correctness.
func Validate(steps []StepConfig) error {
	names := make(map[string]bool)
	outputs := make(map[string]bool)

	for i, step := range steps {
		if step.Name == "" {
			return fmt.Errorf("step %d: name is required", i)
		}
		if names[step.Name] {
			return fmt.Errorf("step %q: duplicate name", step.Name)
		}
		names[step.Name] = true

		if step.Output != "" {
			if outputs[step.Output] {
				return fmt.Errorf("step %q: duplicate output name %q", step.Output, step.Output)
			}
			outputs[step.Output] = true
		}

		if err := validateStep(step, names, outputs); err != nil {
			return fmt.Errorf("step %q: %w", step.Name, err)
		}
	}
	return nil
}

func validateStep(step StepConfig, names, outputs map[string]bool) error {
	switch step.Type {
	case "call":
		if step.Call == nil {
			return fmt.Errorf("type 'call' requires call config")
		}
		if step.Call.Tool == "" {
			return fmt.Errorf("call.tool is required")
		}
	case "map":
		if step.Map == nil {
			return fmt.Errorf("type 'map' requires map config")
		}
		if step.Map.Source == "" {
			return fmt.Errorf("map.source is required")
		}
		if err := Validate(step.Map.Pipeline); err != nil {
			return fmt.Errorf("map.pipeline: %w", err)
		}
	case "transform":
		if step.Transform == nil {
			return fmt.Errorf("type 'transform' requires transform config")
		}
		if step.Transform.Source == "" {
			return fmt.Errorf("transform.source is required")
		}
	case "merge":
		if step.Merge == nil {
			return fmt.Errorf("type 'merge' requires merge config")
		}
		if step.Merge.From == "" || step.Merge.To == "" {
			return fmt.Errorf("merge.from and merge.to are required")
		}
	case "return":
		if step.Return == nil {
			return fmt.Errorf("type 'return' requires return config")
		}
		if step.Return.Source == "" {
			return fmt.Errorf("return.source is required")
		}
	default:
		return fmt.Errorf("unknown step type %q", step.Type)
	}
	return nil
}

// ValidateReferences checks that all {{ }} references in args point to known sources.
// Known sources: input, item, stepName.output, or stepOutput names.
func ValidateReferences(steps []StepConfig) error {
	known := map[string]bool{
		"input": true,
		"item":  true,
	}
	outputNames := make(map[string]string) // output -> step name

	for _, step := range steps {
		if step.Output != "" {
			outputNames[step.Output] = step.Name
		}
		known[step.Name+".output"] = true
	}

	for _, step := range steps {
		refs := extractRefs(step)
		for _, ref := range refs {
			if err := checkRef(ref, step.Name, known, outputNames); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractRefs(step StepConfig) []string {
	var refs []string
	switch step.Type {
	case "call":
		refs = extractMapRefs(step.Call.Args)
	case "map":
		refs = append(refs, extractRefsFromExpr(step.Map.Source)...)
		for _, s := range step.Map.Pipeline {
			refs = append(refs, extractRefs(s)...)
		}
	case "transform":
		refs = append(refs, extractRefsFromExpr(step.Transform.Source)...)
	case "merge":
		refs = append(refs, extractRefsFromExpr(step.Merge.From)...)
		refs = append(refs, extractRefsFromExpr(step.Merge.To)...)
	case "return":
		refs = append(refs, extractRefsFromExpr(step.Return.Source)...)
	}
	return refs
}

func extractMapRefs(m map[string]interface{}) []string {
	var refs []string
	for _, v := range m {
		switch val := v.(type) {
		case string:
			refs = append(refs, extractRefsFromExpr(val)...)
		case map[string]interface{}:
			refs = append(refs, extractMapRefs(val)...)
		case []interface{}:
			for _, item := range val {
				if s, ok := item.(string); ok {
					refs = append(refs, extractRefsFromExpr(s)...)
				}
			}
		}
	}
	return refs
}

func extractRefsFromExpr(expr string) []string {
	matches := refPattern.FindAllStringSubmatch(expr, -1)
	var refs []string
	for _, m := range matches {
		refs = append(refs, strings.TrimSpace(m[1]))
	}
	return refs
}

func checkRef(ref, stepName string, known map[string]bool, outputNames map[string]string) error {
	parts := strings.SplitN(ref, ".", 2)
	root := parts[0]

	// Check if root is a known prefix or output name
	if known[root] {
		return nil
	}
	if _, ok := outputNames[root]; ok {
		return nil
	}
	return fmt.Errorf("step %q: unresolved reference %q (root %q is not input, item, a step output, or a step name)", stepName, ref, root)
}
