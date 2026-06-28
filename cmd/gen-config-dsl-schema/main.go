// schema-gen generates the JSON Schema for mcpgen aggregated tool configuration.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/config"
	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/pipeline"
)

type schema map[string]interface{}

func main() {
	output := flag.String("output", "", "Path to write the schema JSON (default: stdout)")
	flag.Parse()

	s := generate()

	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal schema: %v\n", err)
		os.Exit(1)
	}

	if *output != "" {
		if err := os.WriteFile(*output, b, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write schema: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Schema written to %s\n", *output)
		return
	}
	os.Stdout.Write(b)
	fmt.Println()
}

func generate() schema {
	defs := schema{}

	agg := aggregateToolDef(defs)
	step := stepDef(defs)

	callSpec := schema{
		"type":     "object",
		"required": []string{"tool", "args"},
		"properties": schema{
			"tool":  schema{"type": "string"},
			"parse": schema{"type": "string", "enum": []string{"json"}},
			"args":  schema{"type": "object", "additionalProperties": true},
		},
		"additionalProperties": false,
	}

	jqSpec := schema{
		"type":     "object",
		"required": []string{"expr"},
		"properties": schema{
			"from": schema{"type": "string"},
			"vars": schema{"type": "object", "additionalProperties": true},
			"expr": schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	foreachSpec := schema{
		"type":     "object",
		"required": []string{"in", "as", "pipeline"},
		"properties": schema{
			"in":            schema{"type": "string"},
			"as":            schema{"type": "string"},
			"concurrency":   schema{},
			"preserveOrder": schema{"type": "boolean"},
			"pipeline":      schema{"type": "array", "minItems": 1, "items": ref("#/$defs/Step")},
		},
		"additionalProperties": false,
	}

	returnSpec := schema{
		"type": "object",
		"properties": schema{
			"from": schema{"type": "string"},
			"vars": schema{"type": "object", "additionalProperties": true},
			"expr": schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	emitSpec := schema{
		"type": "object",
		"properties": schema{
			"from": schema{"type": "string"},
			"vars": schema{"type": "object", "additionalProperties": true},
			"expr": schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	requireConfig := schema{
		"type":     "object",
		"required": []string{"nonEmpty"},
		"properties": schema{
			"nonEmpty": schema{"type": "boolean"},
			"message":  schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	defs["AggregatedTool"] = agg
	defs["Step"] = step
	defs["CallSpec"] = callSpec
	defs["JQSpec"] = jqSpec
	defs["ForeachSpec"] = foreachSpec
	defs["ReturnSpec"] = returnSpec
	defs["EmitSpec"] = emitSpec
	defs["RequireConfig"] = requireConfig

	return schema{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$id":     "https://mcpgen/schemas/aggregated-tool-config",
		"title":   "AggregatedToolConfig",
		"description": fmt.Sprintf(
			"Schema for mcpgen aggregated tool pipeline configuration ($HOME/.<binary>/config.yaml). Generated from Go structs: %s",
			sourceFiles(),
		),
		"type":     "object",
		"required": []string{"aggregateTools"},
		"properties": schema{
			"aggregateTools": schema{
				"type":     "array",
				"minItems": 1,
				"items":    ref("#/$defs/AggregatedTool"),
			},
		},
		"$defs": defs,
	}
}

func aggregateToolDef(defs schema) schema {
	return schema{
		"type":     "object",
		"required": []string{"name", "pipeline"},
		"properties": schema{
			"name":        schema{"type": "string"},
			"description": schema{"type": "string"},
			"annotations": schema{"type": "object", "additionalProperties": true},
			"inputSchema": schema{"type": "object", "additionalProperties": true},
			"pipeline": schema{
				"type":     "array",
				"minItems": 1,
				"items":    ref("#/$defs/Step"),
			},
		},
		"additionalProperties": false,
	}
}

func stepDef(defs schema) schema {
	props := schema{
		"id":      schema{"type": "string"},
		"kind":    schema{"type": "string", "enum": stepKinds()},
		"require": ref("#/$defs/RequireConfig"),
		"spec":    schema{},
	}
	step := schema{
		"type":                 "object",
		"required":             []string{"id", "kind", "spec"},
		"properties":           props,
		"additionalProperties": false,
		"allOf":                kindConditionalRequirements(),
	}
	return step
}

func stepKinds() []string {
	return []string{"call", "jq", "foreach", "return", "emit"}
}

func kindConditionalRequirements() []schema {
	kindToDef := map[string]string{
		"call":    "CallSpec",
		"jq":      "JQSpec",
		"foreach": "ForeachSpec",
		"return":  "ReturnSpec",
		"emit":    "EmitSpec",
	}

	var rules []schema
	for _, kind := range stepKinds() {
		refPath := "#/$defs/" + kindToDef[kind]
		rules = append(rules, schema{
			"if":   schema{"properties": schema{"kind": schema{"const": kind}}},
			"then": schema{"properties": schema{"spec": schema{"$ref": refPath}}},
		})
	}
	return rules
}

func ref(path string) schema {
	return schema{"$ref": path}
}

func sourceFiles() string {
	files := []string{
		"config/config.go",
		"pipeline/types.go",
		"pipeline/validator.go",
	}
	return strings.Join(files, ", ")
}

// Compile-time verification.
var (
	_ config.Config
	_ config.AggregatedToolConfig
	_ pipeline.StepConfig
	_ pipeline.StepSpec
	_ pipeline.RequireConfig
)
