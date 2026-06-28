package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"confluence-mcp-v10.2.14/internal/mcpaggregator/pipeline"
)

// Context holds all runtime state for pipeline execution.
type Context struct {
	Input    map[string]interface{}
	Outputs  map[string]interface{} // step id -> output value
	Item     interface{}            // current item in foreach iteration
	ItemName string                 // the "as" name for the current item
}

// NewContext creates a new runtime context.
func NewContext(input map[string]interface{}) *Context {
	return &Context{
		Input:   input,
		Outputs: make(map[string]interface{}),
	}
}

// SetOutput stores a step's output.
func (c *Context) SetOutput(name string, value interface{}) {
	c.Outputs[name] = value
}

// WithItem returns a copy of the context with a different item and item name.
// Outputs are copied so concurrent foreach sub-pipelines are isolated.
func (c *Context) WithItem(item interface{}, asName string) pipeline.StepContext {
	outputs := make(map[string]interface{}, len(c.Outputs)+4)
	for k, v := range c.Outputs {
		outputs[k] = v
	}
	return &Context{
		Input:    c.Input,
		Outputs:  outputs,
		Item:     item,
		ItemName: asName,
	}
}

// ResolvePath resolves a dotted path reference.
// Paths may start with $ (e.g. "$input.field" or "$stepId.field.sub").
// Plain strings without $ are returned as-is.
func (c *Context) ResolvePath(path string) (interface{}, error) {
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "$") {
		return path, nil
	}
	return c.resolveDollarPath(path)
}

// Resolve resolves a value. If the value is a string starting with $, it's treated
// as a reference. Otherwise the value is returned as-is.
func (c *Context) Resolve(expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)
	if !strings.HasPrefix(expr, "$") {
		return expr, nil
	}
	return c.resolveDollarPath(expr)
}

// ResolveMap resolves all values in a map, handling $ references.
func (c *Context) ResolveMap(m map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		resolved, err := c.resolveValue(v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", k, err)
		}
		result[k] = resolved
	}
	return result, nil
}

func (c *Context) resolveValue(v interface{}) (interface{}, error) {
	switch val := v.(type) {
	case string:
		return c.Resolve(val)
	case map[string]interface{}:
		return c.ResolveMap(val)
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			resolved, err := c.resolveValue(item)
			if err != nil {
				return nil, err
			}
			result[i] = resolved
		}
		return result, nil
	default:
		return v, nil
	}
}

// resolveDollarPath resolves a $ref like "$input.field.sub" or "$stepId".
func (c *Context) resolveDollarPath(path string) (interface{}, error) {
	// Strip leading $
	path = path[1:]
	if path == "" {
		return nil, fmt.Errorf("empty $ reference")
	}

	parts := strings.Split(path, ".")
	root := parts[0]

	var current interface{}

	switch root {
	case "input":
		current = c.Input
	case c.ItemName:
		current = c.Item
	default:
		if val, ok := c.Outputs[root]; ok {
			current = val
		} else if val, ok := c.Input[root]; ok && c.ItemName == "" {
			// Allow direct input field access as fallback when not in foreach
			current = val
		} else {
			return nil, fmt.Errorf("unresolved reference: %q (root %q not found in inputs, outputs, or item)", path, root)
		}
	}

	if len(parts) == 1 {
		return current, nil
	}
	return navigatePath(current, parts[1:])
}

func navigatePath(current interface{}, parts []string) (interface{}, error) {
	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("field %q not found at %s", part, strings.Join(parts[:i+1], "."))
			}
		case []interface{}:
			idx := 0
			if _, err := fmt.Sscanf(part, "%d", &idx); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return nil, fmt.Errorf("cannot index array with %q at %s", part, strings.Join(parts[:i+1], "."))
			}
		default:
			b, _ := json.Marshal(current)
			return nil, fmt.Errorf("cannot navigate into %s at %s", string(b), strings.Join(parts[:i], "."))
		}
	}
	return current, nil
}
