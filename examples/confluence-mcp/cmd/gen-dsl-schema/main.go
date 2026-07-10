// Binary gen-dsl-schema generates the JSON Schema (draft 2020-12)
// for this MCP server's configuration by reflecting over the config
// types defined in the serverconfig package.
//
// Usage:
//
//	go run ./cmd/gen-dsl-schema --output dsl-schema.json
//
// When a new config sub-struct is added to serverconfig, it is
// discovered automatically via reflection — no manual wiring needed.
// Only truly new root types (not reachable from existing struct fields)
// need to be added to the Types slice below.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"

	"confluence-mcp/pkg/mcpconfig"
	"confluence-mcp/pkg/mcpvirtual/config"
	"confluence-mcp/pkg/mcpvirtual/pipeline"
	"confluence-mcp/pkg/mcpvirtual/schemagen"
)

func main() {
	output := flag.String("output", "", "Path to write the schema JSON (default: stdout)")
	flag.Parse()

	cfg := schemagen.Config{
		// Root types — the engine recursively discovers all nested struct
		// types. The first type's fields become the root schema properties.
		Types: []reflect.Type{
			reflect.TypeOf(mcpconfig.Config{}),
			reflect.TypeOf(config.VirtualToolConfig{}),
		},
		SchemaTitle:       "MCPServerConfig",
		SchemaID:          "https://mcpfather/schemas/mcp-server-config",
		SchemaDescription: "Schema for MCP server configuration ($HOME/.<binary>/config.yaml). Generated from mcpconfig types.",
		StepKinds:         pipeline.StepKinds,
		Renames: map[string]string{
			"VirtualToolConfig": "VirtualTool",
			"StepConfig":        "Step",
		},
		ExtraRootProps: map[string]schemagen.Schema{
			"virtualTools": {
				"type":     "array",
				"minItems": 1,
				"items":    schemagen.Schema{"$ref": "#/$defs/VirtualTool"},
			},
		},
	}

	schema := schemagen.Generate(cfg)
	b, err := json.MarshalIndent(schema, "", "  ")
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
