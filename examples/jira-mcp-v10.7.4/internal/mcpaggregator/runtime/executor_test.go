package runtime

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"jira-mcp-v10.7.4/internal/mcpaggregator/pipeline"
)

type mockRegistry struct {
	mu          sync.Mutex
	callResults map[string]string
	callCount   map[string]int
}

func (m *mockRegistry) CallTool(ctx context.Context, name string, args map[string]interface{}) (*pipeline.CallToolResult, error) {
	m.mu.Lock()
	m.callCount[name]++
	text := m.callResults[name]
	m.mu.Unlock()
	return &pipeline.CallToolResult{
		Content: []pipeline.ContentItem{{Type: "text", Text: text}},
	}, nil
}

func TestExecutor_SimplePipeline(t *testing.T) {
	reg := &mockRegistry{
		callResults: map[string]string{
			"getData": `{"id": "123", "name": "test"}`,
		},
		callCount: make(map[string]int),
	}

	exec := NewExecutor(reg)
	steps := []pipeline.StepConfig{
		{
			Name: "fetch",
			Type: "call",
			Call: &pipeline.CallConfig{
				Tool: "getData",
				Args: map[string]interface{}{"id": "{{ input.id }}"},
			},
			Output: "data",
		},
		{
			Name: "trim",
			Type: "transform",
			Transform: &pipeline.TransformConfig{
				Source:  "fetch.output",
				Project: []string{"id"},
			},
			Output: "trimmed",
		},
		{
			Name: "done",
			Type: "return",
			Return: &pipeline.ReturnConfig{
				Source: "trimmed.output",
			},
		},
	}

	result, err := exec.Execute(context.Background(), steps, map[string]interface{}{"id": "123"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if result.IsError {
		t.Fatal("result should not be error")
	}

	text := result.Content[0].Text
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("result is not valid JSON: %s", text)
	}
	if data["id"] != "123" {
		t.Errorf("expected id=123, got %v", data["id"])
	}
	if _, ok := data["name"]; ok {
		t.Error("name should have been projected out")
	}

	if reg.callCount["getData"] != 1 {
		t.Errorf("expected 1 call, got %d", reg.callCount["getData"])
	}
}

func TestExecutor_MapPipeline(t *testing.T) {
	reg := &mockRegistry{
		callResults: map[string]string{
			"getItem": `{"value": "enriched"}`,
		},
		callCount: make(map[string]int),
	}

	exec := NewExecutor(reg)
	steps := []pipeline.StepConfig{
		{
			Name: "process",
			Type: "map",
			Map: &pipeline.MapConfig{
				Source: "{{ input.items }}",
				Pipeline: []pipeline.StepConfig{
					{
						Name: "enrich",
						Type: "call",
						Call: &pipeline.CallConfig{
							Tool: "getItem",
							Args: map[string]interface{}{"key": "{{ item.key }}"},
						},
						Output: "itemResult",
					},
					{
						Name: "returnItem",
						Type: "return",
						Return: &pipeline.ReturnConfig{
							Source: "itemResult.output.value",
						},
					},
				},
			},
			Output: "results",
		},
		{
			Name: "done",
			Type: "return",
			Return: &pipeline.ReturnConfig{
				Source: "results.output",
			},
		},
	}

	input := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{"key": "a"},
			map[string]interface{}{"key": "b"},
			map[string]interface{}{"key": "c"},
		},
	}

	result, err := exec.Execute(context.Background(), steps, input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	text := result.Content[0].Text
	var results []interface{}
	if err := json.Unmarshal([]byte(text), &results); err != nil {
		t.Fatalf("map result is not valid JSON array: %s", text)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	if reg.callCount["getItem"] != 3 {
		t.Errorf("expected 3 calls, got %d", reg.callCount["getItem"])
	}
}
