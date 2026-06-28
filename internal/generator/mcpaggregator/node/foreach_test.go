package node

import (
	"context"
	"sync"
	"testing"

	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/pipeline"
)

type mockExecutor struct {
	mu          sync.Mutex
	callResults map[string]string
	callCount   map[string]int
}

func (m *mockExecutor) CallTool(ctx context.Context, name string, args map[string]interface{}) (*pipeline.CallToolResult, error) {
	m.mu.Lock()
	m.callCount[name]++
	text := m.callResults[name]
	m.mu.Unlock()
	return &pipeline.CallToolResult{
		Content: []pipeline.ContentItem{{Type: "text", Text: text}},
	}, nil
}

func (m *mockExecutor) ExecuteStep(ctx context.Context, step *pipeline.StepConfig, rctx pipeline.StepContext) (interface{}, error) {
	switch step.Kind {
	case "call":
		return CallNode(ctx, step, rctx, m)
	case "jq":
		return JQNode(step, rctx)
	case "emit":
		return EmitNode(step, rctx)
	case "return":
		return ReturnValue(step, rctx)
	default:
		panic("unknown kind: " + step.Kind)
	}
}

func TestForeachNode_BasicOrdered(t *testing.T) {
	mock := &mockExecutor{
		callResults: map[string]string{
			"getItem": `{"id": "enriched"}`,
		},
		callCount: make(map[string]int),
	}

	rctx := newMockCtx(nil)
	rctx.SetOutput("items", []interface{}{
		map[string]interface{}{"key": "a"},
		map[string]interface{}{"key": "b"},
		map[string]interface{}{"key": "c"},
	})

	step := &pipeline.StepConfig{
		ID:   "enrich",
		Kind: "foreach",
		Spec: pipeline.StepSpec{
			In:            "$items",
			As:            "item",
			Concurrency:   2,
			PreserveOrder: true,
			Pipeline: []pipeline.StepConfig{
				{
					ID:   "callItem",
					Kind: "call",
					Spec: pipeline.StepSpec{
						Tool: "getItem",
						Args: map[string]interface{}{"key": "$item.key"},
					},
				},
				{
					ID:   "emitResult",
					Kind: "emit",
					Spec: pipeline.StepSpec{From: "$callItem"},
				},
			},
		},
	}

	result, err := ForeachNode(context.Background(), step, rctx, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 results, got %d", len(arr))
	}
	if mock.callCount["getItem"] != 3 {
		t.Errorf("expected 3 calls, got %d", mock.callCount["getItem"])
	}
}

func TestForeachNode_Unordered(t *testing.T) {
	mock := &mockExecutor{
		callResults: map[string]string{
			"getItem": `{"id": "enriched"}`,
		},
		callCount: make(map[string]int),
	}

	rctx := newMockCtx(nil)
	rctx.SetOutput("items", []interface{}{
		map[string]interface{}{"key": "a"},
		map[string]interface{}{"key": "b"},
	})

	step := &pipeline.StepConfig{
		ID:   "enrich",
		Kind: "foreach",
		Spec: pipeline.StepSpec{
			In:            "$items",
			As:            "item",
			Concurrency:   2,
			PreserveOrder: false,
			Pipeline: []pipeline.StepConfig{
				{
					ID:   "callItem",
					Kind: "call",
					Spec: pipeline.StepSpec{
						Tool: "getItem",
						Args: map[string]interface{}{"key": "$item.key"},
					},
				},
				{
					ID:   "emitResult",
					Kind: "emit",
					Spec: pipeline.StepSpec{From: "$callItem"},
				},
			},
		},
	}

	result, err := ForeachNode(context.Background(), step, rctx, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.([]interface{})
	if len(arr) != 2 {
		t.Errorf("expected 2 results, got %d", len(arr))
	}
}

func TestForeachNode_InvalidList(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("data", map[string]interface{}{"not": "an array"})

	step := &pipeline.StepConfig{
		ID:   "bad",
		Kind: "foreach",
		Spec: pipeline.StepSpec{
			In:  "$data",
			As:  "item",
			Pipeline: []pipeline.StepConfig{
				{ID: "e", Kind: "emit", Spec: pipeline.StepSpec{From: "$item"}},
			},
		},
	}

	_, err := ForeachNode(context.Background(), step, rctx, nil)
	if err == nil {
		t.Fatal("expected error for non-array input")
	}
}

func TestForeachNode_ConcurrencyLiteral(t *testing.T) {
	mock := &mockExecutor{
		callResults: map[string]string{
			"echo": `{"ok": true}`,
		},
		callCount: make(map[string]int),
	}

	rctx := newMockCtx(nil)
	items := make([]interface{}, 10)
	for i := 0; i < 10; i++ {
		items[i] = map[string]interface{}{"idx": i}
	}
	rctx.SetOutput("items", items)

	step := &pipeline.StepConfig{
		ID:   "process",
		Kind: "foreach",
		Spec: pipeline.StepSpec{
			In:            "$items",
			As:            "item",
			Concurrency:   5,
			PreserveOrder: true,
			Pipeline: []pipeline.StepConfig{
				{
					ID:   "call",
					Kind: "call",
					Spec: pipeline.StepSpec{
						Tool: "echo",
						Args: map[string]interface{}{"i": "$item.idx"},
					},
				},
				{
					ID:   "emit",
					Kind: "emit",
					Spec: pipeline.StepSpec{From: "$call"},
				},
			},
		},
	}

	result, err := ForeachNode(context.Background(), step, rctx, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.([]interface{})
	if len(arr) != 10 {
		t.Errorf("expected 10 results, got %d", len(arr))
	}
	if mock.callCount["echo"] != 10 {
		t.Errorf("expected 10 calls, got %d", mock.callCount["echo"])
	}
}

func TestForeachNode_DefaultConcurrency(t *testing.T) {
	mock := &mockExecutor{
		callResults: map[string]string{"echo": `{}`},
		callCount:   make(map[string]int),
	}

	rctx := newMockCtx(nil)
	items := make([]interface{}, 3)
	for i := 0; i < 3; i++ {
		items[i] = map[string]interface{}{"idx": i}
	}
	rctx.SetOutput("items", items)

	step := &pipeline.StepConfig{
		ID:   "process",
		Kind: "foreach",
		Spec: pipeline.StepSpec{
			In:  "$items",
			As:  "item",
			Pipeline: []pipeline.StepConfig{
				{
					ID:   "call",
					Kind: "call",
					Spec: pipeline.StepSpec{
						Tool: "echo",
						Args: map[string]interface{}{"i": "$item.idx"},
					},
				},
				{
					ID:   "emit",
					Kind: "emit",
					Spec: pipeline.StepSpec{From: "$call"},
				},
			},
		},
	}

	_, err := ForeachNode(context.Background(), step, rctx, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.callCount["echo"] != 3 {
		t.Errorf("expected 3 calls, got %d", mock.callCount["echo"])
	}
}

func TestForeachNode_NoEmitError(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("items", []interface{}{map[string]interface{}{"x": 1}})

	step := &pipeline.StepConfig{
		ID:   "bad",
		Kind: "foreach",
		Spec: pipeline.StepSpec{
			In:  "$items",
			As:  "item",
			Pipeline: []pipeline.StepConfig{
				{
					ID:   "onlyCall",
					Kind: "call",
					Spec: pipeline.StepSpec{
						Tool: "noexist",
						Args: map[string]interface{}{},
					},
				},
			},
		},
	}

	mock := &mockExecutor{
		callResults: map[string]string{"noexist": `null`},
		callCount:   make(map[string]int),
	}

	_, err := ForeachNode(context.Background(), step, rctx, mock)
	if err == nil {
		t.Fatal("expected error for sub-pipeline without emit")
	}
}

func TestForeachNode_EmptyList(t *testing.T) {
	mock := &mockExecutor{callCount: make(map[string]int)}
	rctx := newMockCtx(nil)
	rctx.SetOutput("items", []interface{}{})

	step := &pipeline.StepConfig{
		ID:   "process",
		Kind: "foreach",
		Spec: pipeline.StepSpec{
			In:  "$items",
			As:  "item",
			Pipeline: []pipeline.StepConfig{
				{
					ID:   "emit",
					Kind: "emit",
					Spec: pipeline.StepSpec{From: "$item"},
				},
			},
		},
	}

	result, err := ForeachNode(context.Background(), step, rctx, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.([]interface{})
	if len(arr) != 0 {
		t.Errorf("expected empty array, got %d items", len(arr))
	}
}

func TestForeachNode_WithJQInSubPipeline(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("items", []interface{}{
		map[string]interface{}{"name": "alice", "age": float64(30), "internal": "x"},
		map[string]interface{}{"name": "bob", "age": float64(25), "internal": "y"},
	})

	step := &pipeline.StepConfig{
		ID:   "clean",
		Kind: "foreach",
		Spec: pipeline.StepSpec{
			In:            "$items",
			As:            "item",
			PreserveOrder: true,
			Pipeline: []pipeline.StepConfig{
				{
					ID:   "transform",
					Kind: "jq",
					Spec: pipeline.StepSpec{
						From: "$item",
						Expr: "del(.internal)",
					},
				},
				{
					ID:   "emitResult",
					Kind: "emit",
					Spec: pipeline.StepSpec{From: "$transform"},
				},
			},
		},
	}

	mock := &mockExecutor{callCount: make(map[string]int)}
	result, err := ForeachNode(context.Background(), step, rctx, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.([]interface{})
	if len(arr) != 2 {
		t.Fatalf("expected 2 results, got %d", len(arr))
	}

	first := arr[0].(map[string]interface{})
	if _, exists := first["internal"]; exists {
		t.Error("internal should be deleted from first item")
	}
	if first["name"] != "alice" {
		t.Errorf("expected alice, got %v", first["name"])
	}
}

func TestResolveConcurrency(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{"nil", nil, 4},
		{"int", 8, 8},
		{"float64", float64(10), 10},
		{"int64", int64(2), 2},
		{"string number", "16", 16},
		{"zero", 0, 0}, // returns 0; caller clamps to 1
	}

	rctx := newMockCtx(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveConcurrency(tt.input, rctx)
			if got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}
