package engine

import (
	"context"
	"fmt"

	"confluence-mcp/internal/mcpvirtual/config"
	"confluence-mcp/internal/mcpvirtual/pipeline"
)

// ToolRegistry provides access to native MCP tools for the virtual tool engine.
type ToolRegistry = pipeline.ToolRegistry

// VirtualToolEntry pairs a tool name/schema with its handler.
type VirtualToolEntry = pipeline.VirtualToolEntry

// Engine manages virtual tools.
type Engine struct {
	config   *config.Config
	registry ToolRegistry
}

// New creates a new Engine from a config path and tool registry.
func New(configPath string, registry ToolRegistry) (*Engine, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load virtual tool config: %w", err)
	}
	return &Engine{config: cfg, registry: registry}, nil
}

// NewFromConfig creates a new Engine from an already-loaded config.
func NewFromConfig(cfg *config.Config, registry ToolRegistry) (*Engine, error) {
	return &Engine{config: cfg, registry: registry}, nil
}

// Tools returns all virtual tool entries for registration with an MCP server.
func (e *Engine) Tools() ([]VirtualToolEntry, error) {
	if e.config == nil || len(e.config.VirtualTools) == 0 {
		return nil, nil
	}

	var entries []VirtualToolEntry
	for _, at := range e.config.VirtualTools {
		entry, err := e.buildTool(at)
		if err != nil {
			return nil, fmt.Errorf("virtual tool %q: %w", at.Name, err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (e *Engine) buildTool(at config.VirtualToolConfig) (VirtualToolEntry, error) {
	if err := pipeline.Validate(at.Pipeline); err != nil {
		return VirtualToolEntry{}, fmt.Errorf("pipeline validation: %w", err)
	}
	if err := pipeline.ValidateReferences(at.Pipeline); err != nil {
		return VirtualToolEntry{}, fmt.Errorf("reference validation: %w", err)
	}

	handler := e.buildHandler(at)

	return VirtualToolEntry{
		Name:        at.Name,
		Description: at.Description,
		InputSchema: at.InputSchema,
		Annotations: at.Annotations,
		Handler:     handler,
	}, nil
}

func (e *Engine) buildHandler(at config.VirtualToolConfig) func(ctx context.Context, args map[string]interface{}) (*pipeline.CallToolResult, error) {
	return func(ctx context.Context, args map[string]interface{}) (*pipeline.CallToolResult, error) {
		if args == nil {
			args = make(map[string]interface{})
		}
		applyDefaults(at.InputSchema, args)
		executor := NewExecutor(e.registry)
		return executor.Execute(ctx, at.Pipeline, args)
	}
}

// applyDefaults merges default values from JSON Schema properties into args.
// Only keys that are missing from args are set (provided values take precedence).
func applyDefaults(schema map[string]interface{}, args map[string]interface{}) {
	props, ok := schema["properties"]
	if !ok {
		return
	}
	propsMap, ok := props.(map[string]interface{})
	if !ok {
		return
	}
	for key, propRaw := range propsMap {
		if _, exists := args[key]; exists {
			continue
		}
		prop, ok := propRaw.(map[string]interface{})
		if !ok {
			continue
		}
		if defaultVal, ok := prop["default"]; ok {
			args[key] = defaultVal
		}
	}
}
