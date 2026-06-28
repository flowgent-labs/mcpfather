package pipeline

import "context"

// StepConfig defines a single step in an aggregated tool pipeline.
type StepConfig struct {
	ID      string         `yaml:"id"`
	Kind    string         `yaml:"kind"` // call, jq, foreach, return, emit
	Spec    StepSpec       `yaml:"spec"`
	Require *RequireConfig `yaml:"require,omitempty"`
}

// StepSpec holds the type-specific configuration for a step.
// Fields are disjoint across kinds — YAML only populates what's present.
type StepSpec struct {
	// call
	Tool  string                 `yaml:"tool,omitempty"`
	Parse string                 `yaml:"parse,omitempty"` // "json" to parse response, empty for raw
	Args  map[string]interface{} `yaml:"args,omitempty"`

	// jq, return, emit (shared)
	From string                 `yaml:"from,omitempty"`
	Vars map[string]interface{} `yaml:"vars,omitempty"`
	Expr string                 `yaml:"expr,omitempty"`

	// foreach
	In            string       `yaml:"in,omitempty"`
	As            string       `yaml:"as,omitempty"`
	Concurrency   interface{}  `yaml:"concurrency,omitempty"` // number or $ref
	PreserveOrder bool         `yaml:"preserveOrder,omitempty"`
	Pipeline      []StepConfig `yaml:"pipeline,omitempty"`
}

// RequireConfig defines post-execution validation on a step's result.
type RequireConfig struct {
	NonEmpty bool   `yaml:"nonEmpty"`
	Message  string `yaml:"message"`
}

// StepContext provides variable resolution and output storage during pipeline execution.
type StepContext interface {
	Resolve(expr string) (interface{}, error)
	ResolvePath(path string) (interface{}, error)
	ResolveMap(m map[string]interface{}) (map[string]interface{}, error)
	SetOutput(name string, value interface{})
	WithItem(item interface{}, asName string) StepContext
}

// StepExecutor executes a single pipeline step within a context.
type StepExecutor interface {
	ExecuteStep(ctx context.Context, step *StepConfig, rctx StepContext) (interface{}, error)
}

// ToolRegistry provides access to native MCP tools.
type ToolRegistry interface {
	CallTool(ctx context.Context, name string, args map[string]interface{}) (*CallToolResult, error)
}

// CallToolResult is a minimal representation of an MCP tool call result.
type CallToolResult struct {
	Content []ContentItem
	IsError bool
}

// ContentItem represents a single content item in a tool result.
type ContentItem struct {
	Type string
	Text string
}

// AggregatedToolEntry pairs a tool name/schema with its handler.
type AggregatedToolEntry struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
	Annotations map[string]interface{}
	Handler     func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error)
}
