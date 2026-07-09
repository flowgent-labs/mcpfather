package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// sortedMap is a map[string]interface{} that marshals keys in sorted order.
type sortedMap struct {
	keys   []string
	values map[string]interface{}
}

func (m sortedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range m.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyJSON, _ := json.Marshal(k)
		buf.Write(keyJSON)
		buf.WriteByte(':')
		valJSON, err := json.Marshal(m.values[k])
		if err != nil {
			return nil, err
		}
		buf.Write(valJSON)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// sortForJSON recursively converts map[string]interface{} to sortedMap
// so that json.Marshal produces deterministic output.
func sortForJSON(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sorted := make(map[string]interface{}, len(val))
		for _, k := range keys {
			sorted[k] = sortForJSON(val[k])
		}
		return sortedMap{keys: keys, values: sorted}
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, item := range val {
			result[i] = sortForJSON(item)
		}
		return result
	default:
		return v
	}
}

// GenerateJSONSchemaDraft7 converts a slice of Arg structs into a JSON Schema Draft 7 string.
// It creates a root object schema with properties for each argument.
func GenerateJSONSchemaDraft7(args []Arg) (string, error) {
	rootSchema := map[string]interface{}{
		"type": "object",
	}

	properties := make(map[string]interface{})
	requiredProperties := []string{}

	for _, arg := range args {
		propSchema, err := buildPropertySchema(arg)
		if err != nil {
			return "", err
		}
		if propSchema == nil {
			continue
		}

		properties[arg.Name] = propSchema

		if arg.Required {
			requiredProperties = append(requiredProperties, arg.Name)
		}
	}

	if len(properties) > 0 {
		rootSchema["properties"] = properties
	}
	if len(requiredProperties) > 0 {
		rootSchema["required"] = requiredProperties
	}

	schemaBytes, err := json.MarshalIndent(sortForJSON(rootSchema), "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON schema: %w", err)
	}

	return string(schemaBytes), nil
}


