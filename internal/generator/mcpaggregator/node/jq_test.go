package node

import (
	"encoding/json"
	"testing"

	"github.com/wl4g-ai/mcpgen/internal/generator/mcpaggregator/pipeline"
)

func TestJQNode_BasicProject(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("data", map[string]interface{}{
		"id": "123", "name": "test", "secret": "xyz",
	})

	step := &pipeline.StepConfig{
		ID:   "trim",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: "$data",
			Expr: "{id, name}",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["id"] != "123" {
		t.Errorf("expected id=123, got %v", m["id"])
	}
	if m["name"] != "test" {
		t.Errorf("expected name=test, got %v", m["name"])
	}
	if _, exists := m["secret"]; exists {
		t.Error("secret should have been projected out")
	}
}

func TestJQNode_FilterSelect(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("items", []interface{}{
		map[string]interface{}{"name": "a", "count": 0},
		map[string]interface{}{"name": "b", "count": 5},
		map[string]interface{}{"name": "c", "count": 3},
	})

	step := &pipeline.StepConfig{
		ID:   "filter",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: "$items",
			Expr: "[.[] | select(.count > 0)]",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array, got %T", result)
	}
	if len(arr) != 2 {
		t.Errorf("expected 2 items, got %d", len(arr))
	}
}

func TestJQNode_WithVars(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("min", float64(4))

	step := &pipeline.StepConfig{
		ID:   "filter",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: `[{"name":"a","level":3},{"name":"b","level":6}]`,
			Vars: map[string]interface{}{"min": "$min"},
			Expr: "[.[] | select(.level >= $min)]",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.([]interface{})
	if len(arr) != 1 {
		t.Errorf("expected 1 item, got %d", len(arr))
	}
}

func TestJQNode_FromLiteral(t *testing.T) {
	rctx := newMockCtx(nil)

	step := &pipeline.StepConfig{
		ID:   "build",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: `{"x": 1, "y": 2}`,
			Expr: "{sum: (.x + .y)}",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.(map[string]interface{})
	sum, _ := m["sum"].(float64)
	if sum != 3 {
		t.Errorf("expected sum=3, got %v", sum)
	}
}

func TestJQNode_Length(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("items", []interface{}{
		map[string]interface{}{"name": "a"},
		map[string]interface{}{"name": "b"},
		map[string]interface{}{"name": "c"},
	})

	step := &pipeline.StepConfig{
		ID:   "count",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: "$items",
			Expr: "length",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	n, ok := result.(int)
	if !ok {
		t.Fatalf("expected int, got %T: %v", result, result)
	}
	if n != 3 {
		t.Errorf("expected 3, got %d", n)
	}
}

func TestJQNode_DelFields(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("data", map[string]interface{}{
		"id": "x", "internal": "secret", "_links": "urls",
	})

	step := &pipeline.StepConfig{
		ID:   "clean",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: "$data",
			Expr: "del(.internal, ._links)",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.(map[string]interface{})
	if _, exists := m["internal"]; exists {
		t.Error("internal should be deleted")
	}
	if _, exists := m["_links"]; exists {
		t.Error("_links should be deleted")
	}
	if m["id"] != "x" {
		t.Error("id should remain")
	}
}

func TestJQNode_NullResult(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("data", map[string]interface{}{})

	step := &pipeline.StepConfig{
		ID:   "empty",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: "$data",
			Expr: "empty",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty result, got %v", result)
	}
}

func TestJQNode_BadExpr(t *testing.T) {
	rctx := newMockCtx(nil)

	step := &pipeline.StepConfig{
		ID:   "bad",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: "$data",
			Expr: "{{{",
		},
	}

	_, err := JQNode(step, rctx)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestJQNode_BadRef(t *testing.T) {
	rctx := newMockCtx(nil)

	step := &pipeline.StepConfig{
		ID:   "bad",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: "$nonexistent",
			Expr: ".",
		},
	}

	_, err := JQNode(step, rctx)
	if err == nil {
		t.Fatal("expected reference error")
	}
}

func TestJQNode_DefaultValue(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("data", map[string]interface{}{"name": "test"})

	step := &pipeline.StepConfig{
		ID:   "def",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: "$data",
			Expr: "{name, level: (.level // 0)}",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.(map[string]interface{})
	lvl, ok := m["level"].(int)
	if !ok || lvl != 0 {
		t.Errorf("expected default level=0, got %v", m["level"])
	}
}

func TestJQNode_MultipleResults(t *testing.T) {
	rctx := newMockCtx(nil)

	step := &pipeline.StepConfig{
		ID:   "iter",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: `[{"a":1},{"a":2},{"a":3}]`,
			Expr: ".[].a",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array for multiple results, got %T", result)
	}
	if len(arr) != 3 {
		t.Errorf("expected 3 results, got %d", len(arr))
	}
}

func TestJQNode_PlainStringFrom(t *testing.T) {
	rctx := newMockCtx(nil)

	step := &pipeline.StepConfig{
		ID:   "wrap",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: "hello",
			Expr: "{message: .}",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.(map[string]interface{})
	if m["message"] != "hello" {
		t.Errorf("expected message=hello, got %v", m["message"])
	}
}

func TestJQNode_ArrayLiteralFrom(t *testing.T) {
	rctx := newMockCtx(nil)

	step := &pipeline.StepConfig{
		ID:   "wrap",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: `[1,2,3]`,
			Expr: "length",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n, ok := result.(int); !ok || n != 3 {
		t.Errorf("expected 3, got %v", result)
	}
}

func TestJQNode_ResolveVarsFromLiteral(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("prefix", "item-")

	step := &pipeline.StepConfig{
		ID:   "build",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: `["a","b"]`,
			Vars: map[string]interface{}{"prefix": "$prefix"},
			Expr: "[.[] | $prefix + .]",
		},
	}

	// gojq variables need special handling - gojq.WithVariables expects variable
	// names like "$varName". resolveJQVars prepends "$".
	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.([]interface{})
	if len(arr) != 2 {
		t.Fatalf("expected 2 items, got %d", len(arr))
	}
	if arr[0] != "item-a" {
		t.Errorf("expected item-a, got %v", arr[0])
	}
	if arr[1] != "item-b" {
		t.Errorf("expected item-b, got %v", arr[1])
	}
}

func TestJQNode_AnyFunction(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("items", []interface{}{
		map[string]interface{}{"ok": false},
		map[string]interface{}{"ok": true},
	})

	step := &pipeline.StepConfig{
		ID:   "check",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: "$items",
			Expr: "any(.ok)",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestResolveFrom_JSONLiteral(t *testing.T) {
	rctx := newMockCtx(nil)

	tests := []struct {
		name string
		from string
	}{
		{"object", `{"key":"value"}`},
		{"array", `[1,2,3]`},
		{"number", "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveFrom(tt.from, rctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.name == "number" {
				if result != "42" {
					t.Errorf("expected 42 as string, got %T %v", result, result)
				}
			} else {
				j, _ := json.Marshal(result)
				t.Logf("resolved %s: %s", tt.from, j)
			}
		})
	}
}

func TestResolveFrom_Empty(t *testing.T) {
	rctx := newMockCtx(nil)
	result, err := resolveFrom("", rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestJQNode_MaxFunction(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("items", []interface{}{
		map[string]interface{}{"score": float64(10)},
		map[string]interface{}{"score": float64(50)},
		map[string]interface{}{"score": float64(30)},
	})

	step := &pipeline.StepConfig{
		ID:   "maxScore",
		Kind: "jq",
		Spec: pipeline.StepSpec{
			From: "$items",
			Expr: "max_by(.score).score",
		},
	}

	result, err := JQNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n, ok := result.(float64); !ok || n != 50 {
		t.Errorf("expected 50 (float64), got %T: %v", result, result)
	}
}
