package engine

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"sonatypeiq-mcp-v1.203.0-01/internal/mcpaggregator/pipeline"
)

type countMockRegistry struct {
	mu          sync.Mutex
	callResults map[string]string
	callCount   map[string]int
}

func (m *countMockRegistry) CallTool(ctx context.Context, name string, args map[string]interface{}) (*pipeline.CallToolResult, error) {
	m.mu.Lock()
	m.callCount[name]++
	text := m.callResults[name]
	m.mu.Unlock()
	return &pipeline.CallToolResult{
		Content: []pipeline.ContentItem{{Type: "text", Text: text}},
	}, nil
}

func TestExecutor_SimplePipeline(t *testing.T) {
	reg := &countMockRegistry{
		callResults: map[string]string{
			"getData": `{"id": "123", "name": "test"}`,
		},
		callCount: make(map[string]int),
	}

	exec := NewExecutor(reg)
	steps := []pipeline.StepConfig{
		{
			ID:   "fetch",
			Kind: "call",
			Spec: pipeline.StepSpec{
				Tool: "getData",
				Args: map[string]interface{}{"id": "$input.id"},
			},
		},
		{
			ID:   "trim",
			Kind: "jq",
			Spec: pipeline.StepSpec{
				From: "$fetch",
				Expr: "{id}",
			},
		},
		{
			ID:   "done",
			Kind: "return",
			Spec: pipeline.StepSpec{From: "$trim"},
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

func TestExecutor_ForeachPipeline(t *testing.T) {
	reg := &countMockRegistry{
		callResults: map[string]string{
			"getItem": `{"value": "enriched"}`,
		},
		callCount: make(map[string]int),
	}

	exec := NewExecutor(reg)
	steps := []pipeline.StepConfig{
		{
			ID:   "process",
			Kind: "foreach",
			Spec: pipeline.StepSpec{
				In:            "$input.items",
				As:            "item",
				Concurrency:   2,
				PreserveOrder: true,
				Pipeline: []pipeline.StepConfig{
					{
						ID:   "enrich",
						Kind: "call",
						Spec: pipeline.StepSpec{
							Tool: "getItem",
							Args: map[string]interface{}{"key": "$item.key"},
						},
					},
					{
						ID:   "emitResult",
						Kind: "emit",
						Spec: pipeline.StepSpec{From: "$enrich.value"},
					},
				},
			},
		},
		{
			ID:   "done",
			Kind: "return",
			Spec: pipeline.StepSpec{From: "$process"},
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
		t.Fatalf("foreach result is not valid JSON array: %s", text)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	if reg.callCount["getItem"] != 3 {
		t.Errorf("expected 3 calls, got %d", reg.callCount["getItem"])
	}
}

func TestExecutor_RequireValidation(t *testing.T) {
	reg := &countMockRegistry{
		callResults: map[string]string{
			"getData": `null`,
		},
		callCount: make(map[string]int),
	}

	exec := NewExecutor(reg)
	steps := []pipeline.StepConfig{
		{
			ID:   "fetch",
			Kind: "call",
			Spec: pipeline.StepSpec{
				Tool: "getData",
				Args: map[string]interface{}{"id": "$input.id"},
			},
			Require: &pipeline.RequireConfig{
				NonEmpty: true,
				Message:  "Data must not be empty",
			},
		},
		{
			ID:   "done",
			Kind: "return",
			Spec: pipeline.StepSpec{From: "$fetch"},
		},
	}

	_, err := exec.Execute(context.Background(), steps, map[string]interface{}{"id": "123"})
	if err == nil {
		t.Fatal("expected require validation error")
	}
}
