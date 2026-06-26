package engine

import (
	"context"
	"testing"

	"confluence-mcp-v10.2.14/internal/mcpaggregator/config"
	"confluence-mcp-v10.2.14/internal/mcpaggregator/pipeline"
)

type mockRegistry struct {
	results map[string]string
}

func (m *mockRegistry) CallTool(ctx context.Context, name string, args map[string]interface{}) (*pipeline.CallToolResult, error) {
	return &pipeline.CallToolResult{
		Content: []pipeline.ContentItem{{Type: "text", Text: m.results[name]}},
	}, nil
}

func TestEngine_BuildTools(t *testing.T) {
	cfg := &config.Config{
		AggregatedTools: []config.AggregatedToolConfig{
			{
				Name:        "my_agg_tool",
				Version:     "1.0",
				Description: "Test aggregated tool",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{"type": "string"},
					},
				},
				Pipeline: []pipeline.StepConfig{
					{
						Name: "fetch",
						Type: "call",
						Call: &pipeline.CallConfig{
							Tool: "native_tool",
							Args: map[string]interface{}{"id": "{{ input.id }}"},
						},
						Output: "data",
					},
					{
						Name: "done",
						Type: "return",
						Return: &pipeline.ReturnConfig{
							Source: "data.output",
						},
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

	engine, err := NewFromConfig(cfg, reg)
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

	// Execute the handler
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
