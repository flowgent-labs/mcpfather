package pipeline

import "context"

// StepConfig defines a single step in an aggregated tool pipeline.
type StepConfig struct {
	Name      string           `yaml:"name"`
	Type      string           `yaml:"type"` // call, map, transform, merge, return
	Call      *CallConfig      `yaml:"call,omitempty"`
	Map       *MapConfig       `yaml:"map,omitempty"`
	Transform *TransformConfig `yaml:"transform,omitempty"`
	Merge     *MergeConfig     `yaml:"merge,omitempty"`
	Return    *ReturnConfig    `yaml:"return,omitempty"`
	Output    string           `yaml:"output,omitempty"`
}

// CallConfig invokes a native MCP tool.
type CallConfig struct {
	Tool string                 `yaml:"tool"`
	Args map[string]interface{} `yaml:"args"`
}

// MapConfig iterates over a source list and executes a sub-pipeline per item.
type MapConfig struct {
	Source   string       `yaml:"source"`
	Pipeline []StepConfig `yaml:"pipeline"`
}

// TransformConfig applies declarative data transformations.
type TransformConfig struct {
	Source  string                 `yaml:"source"`
	Project []string               `yaml:"project,omitempty"`
	Remove  []string               `yaml:"remove,omitempty"`
	Rename  map[string]string      `yaml:"rename,omitempty"`
	Copy    map[string]string      `yaml:"copy,omitempty"`
	Move    map[string]string      `yaml:"move,omitempty"`
	Flatten []string               `yaml:"flatten,omitempty"`
	Default map[string]interface{} `yaml:"default,omitempty"`
}

// MergeConfig merges data from one path into another.
type MergeConfig struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// ReturnConfig returns a value as the final tool result.
type ReturnConfig struct {
	Source string `yaml:"source"`
}

// StepContext provides variable resolution and output storage during pipeline execution.
type StepContext interface {
	Resolve(expr string) (interface{}, error)
	ResolvePath(path string) (interface{}, error)
	ResolveMap(m map[string]interface{}) (map[string]interface{}, error)
	SetOutput(name string, value interface{})
	WithItem(item interface{}) StepContext
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
	Handler     func(ctx context.Context, args map[string]interface{}) (*CallToolResult, error)
}
