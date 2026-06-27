package node

import (
	"fmt"

	"jira-mcp-v10.7.4/internal/mcpaggregator/pipeline"
)

// TransformNode applies declarative transformations to data.
func TransformNode(step *pipeline.StepConfig, rctx pipeline.StepContext) (interface{}, error) {
	cfg := step.Transform

	source, err := rctx.ResolvePath(cfg.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve transform source: %w", err)
	}

	result := deepCopy(source)

	if len(cfg.Project) > 0 {
		result = applyProject(result, cfg.Project)
	}
	if len(cfg.Remove) > 0 {
		result = applyRemove(result, cfg.Remove)
	}
	if len(cfg.Rename) > 0 {
		result = applyRename(result, cfg.Rename)
	}
	if len(cfg.Copy) > 0 {
		result = applyCopy(result, cfg.Copy)
	}
	if len(cfg.Move) > 0 {
		result = applyMove(result, cfg.Move)
	}
	if len(cfg.Flatten) > 0 {
		result = applyFlatten(result, cfg.Flatten)
	}
	if len(cfg.Default) > 0 {
		result = applyDefault(result, cfg.Default)
	}

	return result, nil
}

func applyProject(data interface{}, fields []string) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{})
		for _, f := range fields {
			if val, ok := v[f]; ok {
				out[f] = val
			}
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(v))
		for i, item := range v {
			out[i] = applyProject(item, fields)
		}
		return out
	default:
		return data
	}
}

func applyRemove(data interface{}, fields []string) interface{} {
	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}
	for _, f := range fields {
		delete(m, f)
	}
	return m
}

func applyRename(data interface{}, mapping map[string]string) interface{} {
	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}
	for old, new_ := range mapping {
		if val, exists := m[old]; exists {
			m[new_] = val
			delete(m, old)
		}
	}
	return m
}

func applyCopy(data interface{}, mapping map[string]string) interface{} {
	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}
	for src, dst := range mapping {
		if val, exists := m[src]; exists {
			m[dst] = deepCopy(val)
		}
	}
	return m
}

func applyMove(data interface{}, mapping map[string]string) interface{} {
	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}
	for src, dst := range mapping {
		if val, exists := m[src]; exists {
			m[dst] = val
			delete(m, src)
		}
	}
	return m
}

func applyFlatten(data interface{}, fields []string) interface{} {
	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}
	fieldSet := make(map[string]bool, len(fields))
	for _, f := range fields {
		fieldSet[f] = true
	}
	for key, val := range m {
		if fieldSet[key] {
			if nested, ok := val.(map[string]interface{}); ok {
				for nk, nv := range nested {
					if _, exists := m[nk]; !exists {
						m[nk] = nv
					}
				}
				delete(m, key)
			}
		}
	}
	return m
}

func applyDefault(data interface{}, defaults map[string]interface{}) interface{} {
	m, ok := data.(map[string]interface{})
	if !ok {
		return data
	}
	for key, val := range defaults {
		if _, exists := m[key]; !exists {
			m[key] = val
		}
	}
	return m
}

func deepCopy(src interface{}) interface{} {
	switch v := src.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(v))
		for k, val := range v {
			out[k] = deepCopy(val)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(v))
		for i, val := range v {
			out[i] = deepCopy(val)
		}
		return out
	default:
		return v
	}
}
