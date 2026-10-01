package engine

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/flowgent-labs/mcpfather/pkg/generator/mcpvirtual/config"
	"github.com/flowgent-labs/mcpfather/pkg/generator/mcpvirtual/pipeline"
)

type mockRegistry struct {
	results map[string]string
}

type rejectingHTTPClient struct {
	pool   string
	retry  string
	method string
	err    error
}

func (c *rejectingHTTPClient) Call(context.Context, string, string, string, string, string, map[string]string, map[string]string, interface{}) (*pipeline.HTTPResponse, error) {
	return nil, errors.New("unexpected HTTP call")
}

func (c *rejectingHTTPClient) ValidateHTTPMiddleware(pool, retry, method string) error {
	c.pool, c.retry, c.method = pool, retry, method
	return c.err
}

func (m *mockRegistry) CallTool(ctx context.Context, name string, args map[string]interface{}) (*pipeline.CallToolResult, error) {
	return &pipeline.CallToolResult{
		Content: []pipeline.ContentItem{{Type: "text", Text: m.results[name]}},
	}, nil
}

func TestEngine_BuildTools(t *testing.T) {
	cfg := &config.Config{
		VirtualTools: []config.VirtualToolConfig{
			{
				Name:        "my_agg_tool",
				Description: "Test virtual tool",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{"type": "string"},
					},
				},
				Pipeline: []pipeline.StepConfig{
					{
						ID:   "fetch",
						Kind: "call",
						Spec: pipeline.StepSpec{
							Tool: "native_tool",
							Args: map[string]interface{}{"id": "$input.id"},
						},
					},
					{
						ID:   "done",
						Kind: "return",
						Spec: pipeline.StepSpec{From: "$fetch"},
					},
				},
			},
		},
	}

	reg := &mockRegistry{
		results: map[string]string{
			"native_tool": `{"result": "success"}`,
		},
	}

	engine, err := NewFromConfig(cfg, reg, nil)
	if err != nil {
		t.Fatalf("NewFromConfig failed: %v", err)
	}

	tools, err := engine.Tools()
	if err != nil {
		t.Fatalf("Tools failed: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	entry := tools[0]
	if entry.Name != "my_agg_tool" {
		t.Errorf("name = %q, want %q", entry.Name, "my_agg_tool")
	}
	if entry.Handler == nil {
		t.Fatal("handler is nil")
	}

	result, err := entry.Handler(context.Background(), map[string]interface{}{"id": "123"})
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}
	if result.IsError {
		t.Fatal("handler returned error")
	}
	if len(result.Content) == 0 {
		t.Fatal("handler returned no content")
	}
}

func TestEngine_RejectsInvalidHTTPMiddlewareReference(t *testing.T) {
	wantErr := errors.New("retry policy is not registered")
	client := &rejectingHTTPClient{err: wantErr}
	cfg := &config.Config{VirtualTools: []config.VirtualToolConfig{{
		Name:        "invalid_http_policy",
		Description: "Test HTTP middleware registration validation",
		InputSchema: map[string]interface{}{"type": "object"},
		Pipeline: []pipeline.StepConfig{
			{
				ID:   "fetch",
				Kind: "http",
				Spec: pipeline.StepSpec{
					Upstream: "iq",
					Pool:     "large",
					Retry:    "missing",
					Method:   "POST",
					Path:     "/api/issues",
				},
			},
			{ID: "done", Kind: "return", Spec: pipeline.StepSpec{From: "$fetch.body"}},
		},
	}}}

	engine, err := NewFromConfig(cfg, nil, client)
	if err != nil {
		t.Fatal(err)
	}
	tools, err := engine.Tools()
	if err == nil {
		t.Fatalf("Tools() = %v, nil; want middleware validation error", tools)
	}
	for _, fragment := range []string{"virtual tool \"invalid_http_policy\"", "HTTP middleware validation", "Step \"fetch\"", wantErr.Error()} {
		if !strings.Contains(err.Error(), fragment) {
			t.Errorf("Tools() error %q missing %q", err, fragment)
		}
	}
	if client.pool != "large" || client.retry != "missing" || client.method != "POST" {
		t.Fatalf("validation args = pool:%q retry:%q method:%q", client.pool, client.retry, client.method)
	}
}

func TestEngine_ApplyDefaults(t *testing.T) {
	cfg := &config.Config{
		VirtualTools: []config.VirtualToolConfig{
			{
				Name:        "tool_with_defaults",
				Description: "Test tool with default values",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":    map[string]interface{}{"type": "string"},
						"limit": map[string]interface{}{"type": "integer", "default": 100},
						"debug": map[string]interface{}{"type": "boolean", "default": true},
						"name":  map[string]interface{}{"type": "string", "default": "default_name"},
					},
				},
				Pipeline: []pipeline.StepConfig{
					{
						ID:   "echo",
						Kind: "call",
						Spec: pipeline.StepSpec{
							Tool: "native_echo",
							Args: map[string]interface{}{
								"id":    "$input.id",
								"limit": "$input.limit",
								"debug": "$input.debug",
								"name":  "$input.name",
							},
						},
					},
					{
						ID:   "done",
						Kind: "return",
						Spec: pipeline.StepSpec{From: "$echo"},
					},
				},
			},
		},
	}

	type recordedArgs struct {
		ID    string
		Limit interface{}
		Debug interface{}
		Name  string
	}
	var recorded recordedArgs

	reg := &mockRecordRegistry{
		recordFn: func(name string, args map[string]interface{}) {
			recorded = recordedArgs{
				ID:    args["id"].(string),
				Limit: args["limit"],
				Debug: args["debug"],
				Name:  args["name"].(string),
			}
		},
	}

	engine, err := NewFromConfig(cfg, reg, nil)
	if err != nil {
		t.Fatalf("NewFromConfig failed: %v", err)
	}

	tools, err := engine.Tools()
	if err != nil {
		t.Fatalf("Tools failed: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	// Only provide "id" — defaults should fill in limit, debug, name
	_, err = tools[0].Handler(context.Background(), map[string]interface{}{"id": "abc"})
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if recorded.ID != "abc" {
		t.Errorf("id = %q, want %q", recorded.ID, "abc")
	}
	if recorded.Limit.(int) != 100 {
		t.Errorf("limit = %v, want 100", recorded.Limit)
	}
	if recorded.Debug.(bool) != true {
		t.Errorf("debug = %v, want true", recorded.Debug)
	}
	if recorded.Name != "default_name" {
		t.Errorf("name = %q, want %q", recorded.Name, "default_name")
	}

	// Now provide some defaults — provided values should take precedence
	_, err = tools[0].Handler(context.Background(), map[string]interface{}{
		"id":    "xyz",
		"limit": 50,
	})
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if recorded.ID != "xyz" {
		t.Errorf("id = %q, want %q", recorded.ID, "xyz")
	}
	if recorded.Limit.(int) != 50 {
		t.Errorf("limit = %v, want 50 (provided value should override default)", recorded.Limit)
	}
	if recorded.Debug.(bool) != true {
		t.Errorf("debug = %v, want true (still default since not provided)", recorded.Debug)
	}
	if recorded.Name != "default_name" {
		t.Errorf("name = %q, want %q (still default since not provided)", recorded.Name, "default_name")
	}
}

// mockRecordRegistry records the arguments passed to CallTool.
type mockRecordRegistry struct {
	recordFn func(name string, args map[string]interface{})
}

func (m *mockRecordRegistry) CallTool(ctx context.Context, name string, args map[string]interface{}) (*pipeline.CallToolResult, error) {
	m.recordFn(name, args)
	return &pipeline.CallToolResult{
		Content: []pipeline.ContentItem{{Type: "text", Text: `{"ok": true}`}},
	}, nil
}

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]interface{}
		args     map[string]interface{}
		wantArgs map[string]interface{}
	}{
		{
			name: "no properties",
			schema: map[string]interface{}{
				"type": "object",
			},
			args:     map[string]interface{}{"x": 1},
			wantArgs: map[string]interface{}{"x": 1},
		},
		{
			name: "no properties key at all",
			schema: map[string]interface{}{
				"type": "object",
			},
			args:     map[string]interface{}{},
			wantArgs: map[string]interface{}{},
		},
		{
			name:     "nil schema",
			schema:   nil,
			args:     map[string]interface{}{"x": 1},
			wantArgs: map[string]interface{}{"x": 1},
		},
		{
			name: "missing args get defaults",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer", "default": 100},
					"debug": map[string]interface{}{"type": "boolean", "default": true},
				},
			},
			args: map[string]interface{}{"limit": nil},
			wantArgs: map[string]interface{}{
				"limit": nil, // preserved – nil is a provided value
				"debug": true,
			},
		},
		{
			name: "provided values not overwritten",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer", "default": 100},
				},
			},
			args:     map[string]interface{}{"limit": 50},
			wantArgs: map[string]interface{}{"limit": 50},
		},
		{
			name: "property without default is skipped",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string"},
				},
			},
			args:     map[string]interface{}{},
			wantArgs: map[string]interface{}{},
		},
		{
			name: "string default",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{"type": "string", "default": "hello"},
				},
			},
			args:     map[string]interface{}{},
			wantArgs: map[string]interface{}{"name": "hello"},
		},
		{
			name: "float default",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ratio": map[string]interface{}{"type": "number", "default": 0.75},
				},
			},
			args:     map[string]interface{}{},
			wantArgs: map[string]interface{}{"ratio": 0.75},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyDefaults(tt.schema, tt.args)
		})
	}
}
