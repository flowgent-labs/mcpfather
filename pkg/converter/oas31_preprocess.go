package converter

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// normalizeOAS31 converts OAS 3.1 spec features to OAS 3.0-compatible format
// so kin-openapi can parse them. kin-openapi's Schema struct uses bool for
// ExclusiveMin/ExclusiveMax (OAS 3.0) but OAS 3.1 uses numeric values.
//
// Conversions:
//   - exclusiveMinimum/exclusiveMaximum numeric → bool + min/max
//   - type: ["string", "null"] → type: "string", nullable: true
//   - prefixItems → items (first element as items schema)
//
// Removals (kin-openapi's Schema struct lacks these JSON Schema 2020-12 fields):
//   $schema, $defs, const, $dynamicRef, $dynamicAnchor, $anchor,
//   if/then/else, contains/minContains/maxContains,
//   dependentSchemas, unevaluatedProperties, unevaluatedItems,
//   contentEncoding, contentMediaType, contentSchema, examples.
// Root-level: jsonSchemaDialect.

func normalizeOAS31(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	// Check if this is an OAS 3.1 spec
	if isOpenAPI31(node) {
		// Clean root-level OAS 3.1 keywords
		cleanRootNode(node)
		normalizeSchemas(node)
	}
	return nil
}

// cleanRootNode removes OAS 3.1 root-level fields like jsonSchemaDialect
func cleanRootNode(node *yaml.Node) {
	var root *yaml.Node
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) > 0 {
			root = node.Content[0]
		}
	} else {
		root = node
	}
	if root == nil || root.Kind != yaml.MappingNode {
		return
	}
	rootLevelRemove := map[string]bool{
		"jsonSchemaDialect": true,
	}
	content := root.Content
	j := 0
	for i := 0; i < len(content); i += 2 {
		key := content[i]
		if key.Kind == yaml.ScalarNode && rootLevelRemove[key.Value] {
			continue
		}
		content[j] = content[i]
		content[j+1] = content[i+1]
		j += 2
	}
	root.Content = content[:j]
}

// isOpenAPI31 checks if the document declares openapi: 3.1.x
func isOpenAPI31(root *yaml.Node) bool {
	if root.Kind != yaml.DocumentNode {
		root = &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{root}}
	}
	for _, c := range root.Content {
		if c.Kind == yaml.MappingNode {
			for j := 0; j < len(c.Content); j += 2 {
				if c.Content[j].Value == "openapi" {
					version := c.Content[j+1].Value
					if len(version) >= 3 && version[:3] == "3.1" {
						return true
					}
					return false
				}
			}
		}
	}
	return false
}

// normalizeSchemas recursively finds schema nodes and converts 3.1 features
func normalizeSchemas(node *yaml.Node) {
	if node == nil {
		return
	}
	walkSchemas(node)
}

func walkSchemas(node *yaml.Node) {
	if node == nil {
		return
	}

	if node.Kind == yaml.DocumentNode {
		for _, c := range node.Content {
			walkSchemas(c)
		}
	} else if node.Kind == yaml.MappingNode {
		processSchemaMap(node)
		// Recurse into all values
		for i := 1; i < len(node.Content); i += 2 {
			walkSchemas(node.Content[i])
		}
	} else if node.Kind == yaml.SequenceNode {
		for _, c := range node.Content {
			walkSchemas(c)
		}
	}
}

func processSchemaMap(node *yaml.Node) {
	hasSchemaKeywords := false
	for _, c := range node.Content {
		if c.Kind == yaml.ScalarNode {
			switch c.Value {
			case "type", "properties", "items", "schema", "allOf", "anyOf", "oneOf",
				"exclusiveMinimum", "exclusiveMaximum", "$defs", "prefixItems",
				"additionalProperties", "not", "if", "then", "else", "contains",
				"const", "$schema", "minContains", "maxContains",
				"dependentSchemas", "unevaluatedProperties", "unevaluatedItems",
				"$dynamicAnchor", "$dynamicRef", "$anchor",
				"contentEncoding", "contentMediaType", "contentSchema", "examples":
				hasSchemaKeywords = true
			}
		}
	}
	if !hasSchemaKeywords {
		return
	}

	// 1. Convert numeric exclusiveMinimum/exclusiveMaximum to bool + min/max
	convertExclusiveMinMax(node)

	// 2. Convert type: ["string", "null"] → type: "string", nullable: true
	convertArrayTypeNull(node)

	// 3. Convert prefixItems → items (best-effort: use first element)
	convertPrefixItems(node)

	// 4. Remove OAS 3.1-only keywords that kin-openapi rejects
	//    (includes $defs, const, $schema, dependentSchemas, etc.)
	removeOAS31Keywords(node)
}

// convertExclusiveMinMax converts:
//   exclusiveMinimum: 0  →  exclusiveMinimum: true, minimum: 0
//   exclusiveMinimum: 0.0 → exclusiveMinimum: true, minimum: 0.0
//   exclusiveMaximum: 100 → exclusiveMaximum: true, maximum: 100
// Skips inserting minimum/maximum if the key already exists in the node.
func convertExclusiveMinMax(node *yaml.Node) {
	content := node.Content
	for i := 0; i < len(content); i += 2 {
		key := content[i]
		val := content[i+1]
		if key.Kind != yaml.ScalarNode {
			continue
		}

		switch key.Value {
		case "exclusiveMinimum":
			if (val.Kind == yaml.ScalarNode) && (val.Tag == "!!int" || val.Tag == "!!float") {
				origValue := val.Value
				origTag := val.Tag // preserve !!int or !!float
				val.Value = "true"
				val.Tag = "!!bool"
				val.Kind = yaml.ScalarNode
				// Only insert minimum if not already present
				if !hasKey(node, "minimum") {
					minKey := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "minimum"}
					minVal := &yaml.Node{Kind: yaml.ScalarNode, Tag: origTag, Value: origValue}
					insertAt := i + 2
					newContent := make([]*yaml.Node, 0, len(content)+2)
					newContent = append(newContent, content[:insertAt]...)
					newContent = append(newContent, minKey, minVal)
					newContent = append(newContent, content[insertAt:]...)
					node.Content = newContent
					content = newContent
				}
			}

		case "exclusiveMaximum":
			if (val.Kind == yaml.ScalarNode) && (val.Tag == "!!int" || val.Tag == "!!float") {
				origValue := val.Value
				origTag := val.Tag // preserve !!int or !!float
				val.Value = "true"
				val.Tag = "!!bool"
				val.Kind = yaml.ScalarNode
				// Only insert maximum if not already present
				if !hasKey(node, "maximum") {
					maxKey := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "maximum"}
					maxValNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: origTag, Value: origValue}
					insertAt := i + 2
					newContent := make([]*yaml.Node, 0, len(content)+2)
					newContent = append(newContent, content[:insertAt]...)
					newContent = append(newContent, maxKey, maxValNode)
					newContent = append(newContent, content[insertAt:]...)
					node.Content = newContent
					content = newContent
				}
			}
		}
	}
}

// hasKey checks if a mapping node contains a given key.
func hasKey(node *yaml.Node, key string) bool {
	if node.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Kind == yaml.ScalarNode && node.Content[i].Value == key {
			return true
		}
	}
	return false
}

// convertArrayTypeNull converts:
//   type: ["string", "null"] → type: "string", nullable: true
func convertArrayTypeNull(node *yaml.Node) {
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		if key.Kind != yaml.ScalarNode || key.Value != "type" {
			continue
		}
		if val.Kind == yaml.SequenceNode {
			// Check if "null" is in the type array
			hasNull := false
			var nonNullTypes []string
			for _, t := range val.Content {
				if t.Kind == yaml.ScalarNode {
					if t.Value == "null" {
						hasNull = true
					} else {
						nonNullTypes = append(nonNullTypes, t.Value)
					}
				}
			}
			if hasNull {
				// Set type to first non-null type
				if len(nonNullTypes) > 0 {
					val.Value = nonNullTypes[0]
					val.Kind = yaml.ScalarNode
					val.Tag = "!!str"
					val.Content = nil
				}
				// Add nullable: true
				nulKey := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "nullable"}
				nulVal := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"}
				insertAt := i + 2
				newContent := make([]*yaml.Node, 0, len(node.Content)+2)
				newContent = append(newContent, node.Content[:insertAt]...)
				newContent = append(newContent, nulKey, nulVal)
				newContent = append(newContent, node.Content[insertAt:]...)
				node.Content = newContent
			}
		}
	}
}

// convertPrefixItems converts OAS 3.1 prefixItems → OAS 3.0 items.
// prefixItems is a JSON Schema 2020-12 tuple validation (array of schemas).
// OAS 3.0 items expects a single schema for all elements.
// We take the first element as the items schema.
func convertPrefixItems(node *yaml.Node) {
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		if key.Kind == yaml.ScalarNode && key.Value == "prefixItems" {
			val := node.Content[i+1]
			key.Value = "items"
			if val.Kind == yaml.SequenceNode && len(val.Content) > 0 {
				*val = *val.Content[0]
			}
		}
	}
}

// removeOAS31Keywords removes keywords that kin-openapi doesn't understand
// in OAS 3.0 mode. OAS 3.1 aligns with JSON Schema 2020-12 which introduces
// many keywords absent from kin-openapi's OAS 3.0 Schema struct.
func removeOAS31Keywords(node *yaml.Node) {
	toRemove := map[string]bool{
		// JSON Schema 2020-12 keywords absent from OAS 3.0 Schema
		"const": true,
		// JSON Schema 2020-12 dynamic references
		"$dynamicRef":   true,
		"$dynamicAnchor": true,
		"$anchor":       true,
		"$schema":       true,
		// JSON Schema 2020-12 structural changes
		"$defs":                true,
		"contains":             true,
		"minContains":          true,
		"maxContains":          true,
		"unevaluatedItems":     true,
		"unevaluatedProperties": true,
		// JSON Schema 2020-12 conditional
		"if":   true,
		"then": true,
		"else": true,
		// JSON Schema 2020-12 dependency
		"dependentSchemas": true,
		// JSON Schema 2020-12 content handling
		"contentEncoding":  true,
		"contentMediaType": true,
		"contentSchema":    true,
		// JSON Schema 2020-12 array examples
		"examples": true,
	}
	content := node.Content
	j := 0
	for i := 0; i < len(content); i += 2 {
		key := content[i]
		if key.Kind == yaml.ScalarNode && toRemove[key.Value] {
			continue // skip this key-value pair
		}
		content[j] = content[i]
		content[j+1] = content[i+1]
		j += 2
	}
	node.Content = content[:j]
}

// preprocessSpec loads YAML/JSON into a yaml.Node tree, normalizes OAS 3.1 features,
// then re-encodes to YAML bytes ready for kin-openapi.
func preprocessSpec(data []byte) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := normalizeOAS31(&doc); err != nil {
		return nil, fmt.Errorf("failed to normalize OAS 3.1 spec: %w", err)
	}

	return yaml.Marshal(&doc)
}
