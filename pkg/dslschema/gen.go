// Package dslschema generates a JSON Schema (draft 2020-12) for
// mcpfather MCP server configuration, derived from the Go struct
// definitions in config.templ, mcpvirtual/config, and mcpvirtual/pipeline.
package dslschema

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Schema is a JSON Schema document represented as a generic map.
type Schema map[string]interface{}

// Generate returns the full MCP server config JSON Schema.
func Generate() Schema {
	defs := Schema{}

	// ---- runtime ----
	upstreamTool := Schema{
		"type":     "object",
		"required": []string{"name"},
		"properties": Schema{
			"name": Schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	upstream := Schema{
		"type":     "object",
		"required": []string{"endpoint"},
		"properties": Schema{
			"endpoint":                        Schema{"type": "string"},
			"enable_mcp_session_in_forwarding": Schema{"type": "boolean"},
			"tools": Schema{
				"type":  "array",
				"items": ref("#/$defs/UpstreamToolConfig"),
			},
		},
		"additionalProperties": false,
	}

	runtime := Schema{
		"type":     "object",
		"properties": Schema{
			"download_dir":       Schema{"type": "string"},
			"log_authorization": Schema{"type": "boolean"},
		},
		"additionalProperties": false,
	}

	// ---- mgmt ----
	pprof := Schema{
		"type": "object",
		"properties": Schema{
			"enabled":     Schema{"type": "boolean"},
			"server_bind": Schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	otel := Schema{
		"type": "object",
		"properties": Schema{
			"enabled":     Schema{"type": "boolean"},
			"endpoint":    Schema{"type": "string"},
			"protocol":    Schema{"type": "string"},
			"timeout":     Schema{"type": "integer"},
			"sample_rate": Schema{"type": "number"},
		},
		"additionalProperties": false,
	}

	metrics := Schema{
		"type": "object",
		"properties": Schema{
			"enabled":              Schema{"type": "boolean"},
			"prometheus":           Schema{"type": "boolean"},
			"export_interval":      Schema{"type": "string"},
			"histogram_boundaries": Schema{"type": "object", "additionalProperties": Schema{"type": "array", "items": Schema{"type": "number"}}},
			"labels":               Schema{"type": "object", "additionalProperties": Schema{"type": "string"}},
		},
		"additionalProperties": false,
	}

	mgmt := Schema{
		"type": "object",
		"properties": Schema{
			"enabled": Schema{"type": "boolean"},
			"host":    Schema{"type": "string"},
			"port":    Schema{"type": "integer"},
			"pprof":   ref("#/$defs/PprofConfig"),
			"otel":    ref("#/$defs/OtelConfig"),
			"metrics": ref("#/$defs/MetricsConfig"),
		},
		"additionalProperties": false,
	}

	// ---- auth ----
	frontendOIDC := Schema{
		"type": "object",
		"properties": Schema{
			"enabled":  Schema{"type": "boolean"},
			"issuer":   Schema{"type": "string"},
			"jwks_uri": Schema{"type": "string"},
			"audience": Schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	frontend := Schema{
		"type": "object",
		"properties": Schema{
			"oidc": ref("#/$defs/FrontendOIDCConfig"),
		},
		"additionalProperties": false,
	}

	backendOIDC := Schema{
		"type": "object",
		"properties": Schema{
			"enabled":       Schema{"type": "boolean"},
			"issuer":        Schema{"type": "string"},
			"client_id":     Schema{"type": "string"},
			"client_secret": Schema{"type": "string"},
			"scopes":        Schema{"type": "string"},
			"grant_type":    Schema{"type": "string"},
			"token_url":     Schema{"type": "string"},
			"username":      Schema{"type": "string"},
			"password":      Schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	backendLDAP := Schema{
		"type": "object",
		"properties": Schema{
			"enabled":              Schema{"type": "boolean"},
			"url":                  Schema{"type": "string"},
			"base_dn":              Schema{"type": "string"},
			"bind_dn":              Schema{"type": "string"},
			"bind_password":        Schema{"type": "string"},
			"insecure_skip_verify": Schema{"type": "boolean"},
			"timeout":              Schema{"type": "integer"},
		},
		"additionalProperties": false,
	}

	backendStatic := Schema{
		"type": "object",
		"properties": Schema{
			"bearer_token":      Schema{"type": "string"},
			"bearer_token_file": Schema{"type": "string"},
			"cookie_token":      Schema{"type": "string"},
			"cookie_token_file": Schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	backend := Schema{
		"type": "object",
		"properties": Schema{
			"oidc":   ref("#/$defs/BackendOIDCConfig"),
			"ldap":   ref("#/$defs/BackendLDAPConfig"),
			"static": ref("#/$defs/BackendStaticConfig"),
		},
		"additionalProperties": false,
	}

	auth := Schema{
		"type": "object",
		"properties": Schema{
			"frontend": ref("#/$defs/FrontendAuthConfig"),
			"backend":  ref("#/$defs/BackendAuthConfig"),
		},
		"additionalProperties": false,
	}

	// ---- tools ----
	toolsExpose := Schema{
		"type": "object",
		"properties": Schema{
			"register_all_tools_by_default": Schema{"type": "boolean"},
			"includes":                      Schema{"type": "array", "items": Schema{"type": "string"}},
			"excludes":                      Schema{"type": "array", "items": Schema{"type": "string"}},
		},
		"additionalProperties": false,
	}

	tools := Schema{
		"type": "object",
		"properties": Schema{
			"expose": ref("#/$defs/ToolsExposeConfig"),
		},
		"additionalProperties": false,
	}

	// ---- virtualTools (from existing DSL schema) ----
	vt := virtualToolDef(defs)
	step := stepDef(defs)

	callSpec := Schema{
		"type":     "object",
		"required": []string{"tool", "args"},
		"properties": Schema{
			"tool":  Schema{"type": "string"},
			"parse": Schema{"type": "string", "enum": []string{"json"}},
			"args":  Schema{"type": "object", "additionalProperties": true},
		},
		"additionalProperties": false,
	}

	jqSpec := Schema{
		"type":     "object",
		"required": []string{"expr"},
		"properties": Schema{
			"from": Schema{"type": "string"},
			"vars": Schema{"type": "object", "additionalProperties": true},
			"expr": Schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	foreachSpec := Schema{
		"type":     "object",
		"required": []string{"in", "as", "pipeline"},
		"properties": Schema{
			"in":            Schema{"type": "string"},
			"as":            Schema{"type": "string"},
			"concurrency":   Schema{},
			"preserveOrder": Schema{"type": "boolean"},
			"onMissing":     Schema{"type": "string", "enum": []string{"skip", "error"}},
			"pipeline":      Schema{"type": "array", "minItems": 1, "items": ref("#/$defs/Step")},
		},
		"additionalProperties": false,
	}

	returnSpec := Schema{
		"type": "object",
		"properties": Schema{
			"from": Schema{"type": "string"},
			"vars": Schema{"type": "object", "additionalProperties": true},
			"expr": Schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	emitSpec := Schema{
		"type": "object",
		"properties": Schema{
			"from": Schema{"type": "string"},
			"vars": Schema{"type": "object", "additionalProperties": true},
			"expr": Schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	requireConfig := Schema{
		"type":     "object",
		"required": []string{"nonEmpty"},
		"properties": Schema{
			"nonEmpty": Schema{"type": "boolean"},
			"field":    Schema{"type": "string"},
			"message":  Schema{"type": "string"},
		},
		"additionalProperties": false,
	}

	// Register all defs
	defs["UpstreamToolConfig"] = upstreamTool
	defs["UpstreamConfig"] = upstream
	defs["RuntimeConfig"] = runtime
	defs["PprofConfig"] = pprof
	defs["OtelConfig"] = otel
	defs["MetricsConfig"] = metrics
	defs["MgmtConfig"] = mgmt
	defs["FrontendOIDCConfig"] = frontendOIDC
	defs["FrontendAuthConfig"] = frontend
	defs["BackendOIDCConfig"] = backendOIDC
	defs["BackendLDAPConfig"] = backendLDAP
	defs["BackendStaticConfig"] = backendStatic
	defs["BackendAuthConfig"] = backend
	defs["AuthConfig"] = auth
	defs["ToolsExposeConfig"] = toolsExpose
	defs["ToolsConfig"] = tools
	defs["VirtualTool"] = vt
	defs["Step"] = step
	defs["CallSpec"] = callSpec
	defs["JQSpec"] = jqSpec
	defs["ForeachSpec"] = foreachSpec
	defs["ReturnSpec"] = returnSpec
	defs["EmitSpec"] = emitSpec
	defs["RequireConfig"] = requireConfig

	return Schema{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"$id":     "https://mcpfather/schemas/mcp-server-config",
		"title":   "MCPServerConfig",
		"description": fmt.Sprintf(
			"Schema for mcpfather MCP server configuration ($HOME/.<binary>/config.yaml). Generated from Go structs: %s",
			sourceFiles(),
		),
		"type": "object",
		"properties": Schema{
			"upstream":     ref("#/$defs/UpstreamConfig"),
			"runtime":      ref("#/$defs/RuntimeConfig"),
			"mgmt":         ref("#/$defs/MgmtConfig"),
			"auth":         ref("#/$defs/AuthConfig"),
			"tools":        ref("#/$defs/ToolsConfig"),
			"virtualTools": Schema{
				"type":     "array",
				"minItems": 1,
				"items":    ref("#/$defs/VirtualTool"),
			},
		},
		"additionalProperties": true,
		"$defs":                defs,
	}
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

func virtualToolDef(defs Schema) Schema {
	return Schema{
		"type":     "object",
		"required": []string{"name", "pipeline"},
		"properties": Schema{
			"name":        Schema{"type": "string"},
			"description": Schema{"type": "string"},
			"enabled":     Schema{"type": "boolean"},
			"annotations": Schema{"type": "object", "additionalProperties": true},
			"inputSchema": Schema{"type": "object", "additionalProperties": true},
			"pipeline": Schema{
				"type":     "array",
				"minItems": 1,
				"items":    ref("#/$defs/Step"),
			},
		},
		"additionalProperties": false,
	}
}

func stepDef(defs Schema) Schema {
	step := Schema{
		"type":     "object",
		"required": []string{"id", "kind", "spec"},
		"properties": Schema{
			"id":      Schema{"type": "string"},
			"kind":    Schema{"type": "string", "enum": stepKinds()},
			"require": ref("#/$defs/RequireConfig"),
			"spec":    Schema{},
		},
		"additionalProperties": false,
		"allOf":                kindConditionalRequirements(),
	}
	return step
}

func stepKinds() []string {
	return []string{"call", "jq", "foreach", "return", "emit"}
}

func kindConditionalRequirements() []Schema {
	kindToDef := map[string]string{
		"call":    "CallSpec",
		"jq":      "JQSpec",
		"foreach": "ForeachSpec",
		"return":  "ReturnSpec",
		"emit":    "EmitSpec",
	}

	var rules []Schema
	for _, kind := range stepKinds() {
		refPath := "#/$defs/" + kindToDef[kind]
		rules = append(rules, Schema{
			"if":   Schema{"properties": Schema{"kind": Schema{"const": kind}}},
			"then": Schema{"properties": Schema{"spec": Schema{"$ref": refPath}}},
		})
	}
	return rules
}

func ref(path string) Schema {
	return Schema{"$ref": path}
}

func sourceFiles() string {
	files := []string{
		"config.templ",
		"config/config.go",
		"pipeline/types.go",
		"pipeline/validator.go",
	}
	return strings.Join(files, ", ")
}
