// Package dslschema generates a JSON Schema (draft 2020-12) for
// mcpfather MCP server configuration. It delegates schema generation
// to the reflection engine in schemagen, reflecting over the canonical
// config types defined in mcpconfig.
//
// Generated MCP server projects use cmd/gen-dsl-schema directly via
// their own copy of the mcpconfig types + schemagen engine.
package dslschema

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/flowgent-labs/mcpfather/pkg/generator/mcpconfig"
	"github.com/flowgent-labs/mcpfather/pkg/generator/mcpvirtual/config"
	"github.com/flowgent-labs/mcpfather/pkg/generator/mcpvirtual/pipeline"
	"github.com/flowgent-labs/mcpfather/pkg/generator/mcpvirtual/schemagen"
)

// Generate uses the schemagen reflection engine over the canonical
// config types in serverconfig (no mirror structs).
func Generate() schemagen.Schema {
	return schemagen.Generate(schemagen.Config{
		Types: []reflect.Type{
			reflect.TypeOf(mcpconfig.Config{}),
			reflect.TypeOf(config.VirtualToolConfig{}),
		},
		SchemaTitle:       "MCPServerConfig",
		SchemaID:          "https://mcpfather/schemas/mcp-server-config",
		SchemaDescription: "Schema for mcpfather MCP server configuration ($HOME/.<binary>/config.yaml). Generated from Go structs in mcpconfig + mcpvirtual/config + mcpvirtual/pipeline.",
		Renames: map[string]string{
			"VirtualToolConfig": "VirtualTool",
			"StepConfig":        "Step",
		},
		StepKinds: pipeline.StepKinds,
		ExtraRootProps: map[string]schemagen.Schema{
			"virtualTools": {
				"type":     "array",
				"minItems": 1,
				"items":    schemagen.Schema{"$ref": "#/$defs/VirtualTool"},
			},
		},
	})
}

// Write writes the JSON Schema to a file.
func Write(outputPath string) error {
	b, err := Encode()
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, b, 0644)
}

// Encode marshals the JSON Schema to indented JSON bytes.
func Encode() ([]byte, error) {
	s := Generate()
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	return b, nil
}
