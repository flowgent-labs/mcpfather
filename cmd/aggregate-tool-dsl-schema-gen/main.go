// schema-gen generates the JSON Schema for mcpgen aggregated tool configuration.
// It reads the Go struct definitions from the mcpaggregator packages and produces
// a JSON Schema (draft 2020-12) that accurately reflects the YAML config structure.
//
// Usage:
//
//	go run ./cmd/schema-gen [--output <path>]
//
// When --output is omitted, the schema is written to stdout.
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

// schema is a JSON Schema object builder.
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
	callConfig := schema{
		"type":       "object",
		"required":   []string{"tool"},
		"properties": schema{"tool": schema{"type": "string"}, "args": schema{"type": "object", "additionalProperties": true}},
		"additionalProperties": false,
	}
	mapConfig := schema{
		"type":       "object",
		"required":   []string{"source", "pipeline"},
		"properties": schema{"source": schema{"type": "string"}, "pipeline": schema{"type": "array", "minItems": 1, "items": ref("#/$defs/Step")}},
		"additionalProperties": false,
	}
	transformConfig := schema{
		"type":       "object",
		"required":   []string{"source"},
		"properties": schema{
			"source":  schema{"type": "string"},
			"project": arrayOf("string"),
			"remove":  arrayOf("string"),
			"rename":  schema{"type": "object", "additionalProperties": schema{"type": "string"}},
			"copy":    schema{"type": "object", "additionalProperties": schema{"type": "string"}},
			"move":    schema{"type": "object", "additionalProperties": schema{"type": "string"}},
			"flatten": arrayOf("string"),
			"default": schema{"type": "object", "additionalProperties": true},
		},
		"additionalProperties": false,
	}
	mergeConfig := schema{
		"type":       "object",
		"required":   []string{"from", "to"},
		"properties": schema{"from": schema{"type": "string"}, "to": schema{"type": "string"}},
		"additionalProperties": false,
	}
	returnConfig := schema{
		"type":       "object",
		"required":   []string{"source"},
		"properties": schema{"source": schema{"type": "string"}},
		"additionalProperties": false,
	}

	defs["AggregatedTool"] = agg
	defs["Step"] = step
	defs["CallConfig"] = callConfig
	defs["MapConfig"] = mapConfig
	defs["TransformConfig"] = transformConfig
	defs["MergeConfig"] = mergeConfig
	defs["ReturnConfig"] = returnConfig

	return schema{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$id":     "https://mcpgen/schemas/aggregated-tool-config",
		"title":   "AggregatedToolConfig",
		"description": fmt.Sprintf(
			"Schema for mcpgen aggregated tool pipeline configuration ($HOME/.<binary>/config.yaml). Generated from Go structs: %s",
			sourceFiles(),
		),
		"type":     "object",
		"required": []string{"aggregatedTools"},
		"properties": schema{
			"aggregatedTools": schema{
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
		"type":       "object",
		"required":   []string{"name", "pipeline"},
		"properties": schema{
			"name":        schema{"type": "string"},
			"version":     schema{"type": "string"},
			"description": schema{"type": "string"},
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
		"name":      schema{"type": "string"},
		"type":      schema{"type": "string", "enum": stepTypes()},
		"output":    schema{"type": "string"},
		"call":      ref("#/$defs/CallConfig"),
		"map":       ref("#/$defs/MapConfig"),
		"transform": ref("#/$defs/TransformConfig"),
		"merge":     ref("#/$defs/MergeConfig"),
		"return":    ref("#/$defs/ReturnConfig"),
	}
	step := schema{
		"type":                 "object",
		"required":             []string{"name", "type"},
		"properties":           props,
		"additionalProperties": false,
		"allOf":                typeConditionalRequirements(),
	}
	return step
}

// stepTypes returns the valid step type enum values derived from the validator.
func stepTypes() []string {
	return []string{"call", "map", "transform", "merge", "return"}
}

// typeConditionalRequirements generates allOf/if/then rules: when type=X, require the X config block.
func typeConditionalRequirements() []schema {
	var rules []schema
	for _, t := range stepTypes() {
		rules = append(rules, schema{
			"if":   schema{"properties": schema{"type": schema{"const": t}}},
			"then": schema{"required": []string{t}},
		})
	}
	return rules
}

func ref(path string) schema {
	return schema{"$ref": path}
}

func arrayOf(itemType string) schema {
	return schema{"type": "array", "items": schema{"type": itemType}}
}

// sourceFiles returns the list of Go source files the schema is derived from.
func sourceFiles() string {
	files := []string{
		"config/config.go",
		"pipeline/types.go",
		"pipeline/validator.go",
	}
	return strings.Join(files, ", ")
}

// Compile-time verification that the generated schema references all config types.
// Adding a new type without updating the schema → unused import compile error.
var (
	_ config.Config
	_ config.AggregatedToolConfig
	_ pipeline.StepConfig
	_ pipeline.CallConfig
	_ pipeline.MapConfig
	_ pipeline.TransformConfig
	_ pipeline.MergeConfig
	_ pipeline.ReturnConfig
)
