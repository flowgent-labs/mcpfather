package engine

import (
	"context"
	"fmt"

	"jira-mcp-v10.7.4/internal/mcpaggregator/config"
	"jira-mcp-v10.7.4/internal/mcpaggregator/pipeline"
	"jira-mcp-v10.7.4/internal/mcpaggregator/runtime"
)

// ToolRegistry provides access to native MCP tools for the aggregated tool engine.
type ToolRegistry = pipeline.ToolRegistry

// AggregatedToolEntry pairs a tool name/schema with its handler.
type AggregatedToolEntry = pipeline.AggregatedToolEntry

// Engine manages aggregated tools.
type Engine struct {
	config   *config.Config
	registry ToolRegistry
}

// New creates a new Engine from a config path and tool registry.
func New(configPath string, registry ToolRegistry) (*Engine, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load aggregated tool config: %w", err)
	}
	return &Engine{config: cfg, registry: registry}, nil
}

// NewFromConfig creates a new Engine from an already-loaded config.
func NewFromConfig(cfg *config.Config, registry ToolRegistry) (*Engine, error) {
	return &Engine{config: cfg, registry: registry}, nil
}

// Tools returns all aggregated tool entries for registration with an MCP server.
// Returns an error if any tool's configuration is invalid.
func (e *Engine) Tools() ([]AggregatedToolEntry, error) {
	if e.config == nil || len(e.config.AggregatedTools) == 0 {
		return nil, nil
	}

	var entries []AggregatedToolEntry
	for _, at := range e.config.AggregatedTools {
		entry, err := e.buildTool(at)
		if err != nil {
			return nil, fmt.Errorf("aggregated tool %q: %w", at.Name, err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (e *Engine) buildTool(at config.AggregatedToolConfig) (AggregatedToolEntry, error) {
	if err := pipeline.Validate(at.Pipeline); err != nil {
		return AggregatedToolEntry{}, fmt.Errorf("pipeline validation: %w", err)
	}
	if err := pipeline.ValidateReferences(at.Pipeline); err != nil {
		return AggregatedToolEntry{}, fmt.Errorf("reference validation: %w", err)
	}

	handler := e.buildHandler(at)

	return AggregatedToolEntry{
		Name:        at.Name,
		Description: at.Description,
		InputSchema: at.InputSchema,
		Handler:     handler,
	}, nil
}

func (e *Engine) buildHandler(at config.AggregatedToolConfig) func(ctx context.Context, args map[string]interface{}) (*pipeline.CallToolResult, error) {
	return func(ctx context.Context, args map[string]interface{}) (*pipeline.CallToolResult, error) {
		if args == nil {
			args = make(map[string]interface{})
		}
		executor := runtime.NewExecutor(e.registry)
		return executor.Execute(ctx, at.Pipeline, args)
	}
}
