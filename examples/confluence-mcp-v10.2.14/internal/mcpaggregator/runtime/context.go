package runtime

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"confluence-mcp-v10.2.14/internal/mcpaggregator/pipeline"
)

var refPattern = regexp.MustCompile(`\{\{\s*([^}]+)\s*\}\}`)

// Context holds all runtime state for pipeline execution.
type Context struct {
	Input   map[string]interface{}
	Outputs map[string]interface{} // step name -> output value
	Item    interface{}            // current item in map iteration
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

// GetOutput retrieves a step's output.
func (c *Context) GetOutput(name string) (interface{}, bool) {
	v, ok := c.Outputs[name]
	return v, ok
}

// WithItem returns a copy of the context with a different item.
// Outputs are copied so concurrent map sub-pipelines are isolated.
func (c *Context) WithItem(item interface{}) pipeline.StepContext {
	outputs := make(map[string]interface{}, len(c.Outputs)+4)
	for k, v := range c.Outputs {
		outputs[k] = v
	}
	return &Context{
		Input:   c.Input,
		Outputs: outputs,
		Item:    item,
	}
}

// ResolvePath resolves a dotted path reference (without needing {{ }} wrappers).
// Supports: input.field, stepName.output.field, item.field, stepName.field
func (c *Context) ResolvePath(path string) (interface{}, error) {
	// Strip {{ }} if present
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "{{") && strings.HasSuffix(path, "}}") {
		path = strings.TrimSpace(path[2 : len(path)-2])
	}
	return c.resolvePath(path)
}

// Resolve resolves a reference string like "{{ input.field }}" or "stepName.output.field".
// If the value is a plain string without {{ }}, it is returned as-is.
func (c *Context) Resolve(expr string) (interface{}, error) {
	// Check if the whole expression is a single reference
	trimmed := strings.TrimSpace(expr)
	if refPattern.MatchString(trimmed) {
		matches := refPattern.FindAllStringSubmatch(trimmed, -1)
		if len(matches) == 1 && matches[0][0] == trimmed {
			return c.resolvePath(strings.TrimSpace(matches[0][1]))
		}
		// Multiple refs or mixed text: substitute all refs as strings
		result := refPattern.ReplaceAllStringFunc(trimmed, func(match string) string {
			sub := refPattern.FindStringSubmatch(match)
			if sub == nil {
				return match
			}
			val, err := c.resolvePath(strings.TrimSpace(sub[1]))
			if err != nil {
				return match
			}
			return fmt.Sprintf("%v", val)
		})
		return result, nil
	}
	return expr, nil
}

// ResolveString resolves a reference and returns it as a string.
func (c *Context) ResolveString(expr string) (string, error) {
	val, err := c.Resolve(expr)
	if err != nil {
		return "", err
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	return fmt.Sprintf("%v", val), nil
}

// ResolveMap resolves all values in a map, handling {{ }} references.
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

func (c *Context) resolvePath(path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	root := parts[0]
	var current interface{}
	skip := 1 // how many leading parts we've consumed

	switch root {
	case "input":
		current = c.Input
	case "item":
		current = c.Item
	default:
		if val, ok := c.Outputs[root]; ok {
			current = val
			// If next segment is "output", it's syntactic sugar — skip it.
			if len(parts) >= 2 && parts[1] == "output" {
				skip = 2
			}
		} else {
			return nil, fmt.Errorf("unresolved reference: %q (root %q not found)", path, root)
		}
	}

	if skip >= len(parts) {
		return current, nil
	}
	return navigatePath(current, parts[skip:])
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
			// Try to parse as index
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
