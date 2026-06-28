package node

import (
	"fmt"
	"strings"

	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/pipeline"
)

// mockCtx is a lightweight StepContext implementation for node tests.
type mockCtx struct {
	input    map[string]interface{}
	outputs  map[string]interface{}
	item     interface{}
	itemName string
}

func newMockCtx(input map[string]interface{}) *mockCtx {
	return &mockCtx{
		input:   input,
		outputs: make(map[string]interface{}),
	}
}

func (c *mockCtx) SetOutput(name string, value interface{}) {
	c.outputs[name] = value
}

func (c *mockCtx) WithItem(item interface{}, asName string) pipeline.StepContext {
	outputs := make(map[string]interface{}, len(c.outputs)+4)
	for k, v := range c.outputs {
		outputs[k] = v
	}
	return &mockCtx{
		input:    c.input,
		outputs:  outputs,
		item:     item,
		itemName: asName,
	}
}

func (c *mockCtx) Resolve(expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)
	if !strings.HasPrefix(expr, "$") {
		return expr, nil
	}
	return c.resolveDollarPath(expr)
}

func (c *mockCtx) ResolvePath(path string) (interface{}, error) {
	path = strings.TrimSpace(path)
	if !strings.HasPrefix(path, "$") {
		return path, nil
	}
	return c.resolveDollarPath(path)
}

func (c *mockCtx) ResolveMap(m map[string]interface{}) (map[string]interface{}, error) {
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

func (c *mockCtx) resolveValue(v interface{}) (interface{}, error) {
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

func (c *mockCtx) resolveDollarPath(path string) (interface{}, error) {
	path = path[1:] // strip $
	if path == "" {
		return nil, fmt.Errorf("empty $ reference")
	}

	parts := strings.Split(path, ".")
	root := parts[0]

	var current interface{}

	switch root {
	case "input":
		current = c.input
	case c.itemName:
		current = c.item
	default:
		if val, ok := c.outputs[root]; ok {
			current = val
		} else if val, ok := c.input[root]; ok && c.itemName == "" {
			current = val
		} else {
			return nil, fmt.Errorf("unresolved reference: %q", root)
		}
	}

	if len(parts) == 1 {
		return current, nil
	}
	return navigateFields(current, parts[1:])
}

func navigateFields(current interface{}, parts []string) (interface{}, error) {
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("field %q not found", part)
			}
		default:
			return nil, fmt.Errorf("cannot navigate into %T", current)
		}
	}
	return current, nil
}
