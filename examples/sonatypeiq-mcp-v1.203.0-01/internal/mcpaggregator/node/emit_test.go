package node

import (
	"testing"

	"sonatypeiq-mcp-v1.203.0-01/internal/mcpaggregator/pipeline"
)

func TestEmitNode_SimpleFrom(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("step1", map[string]interface{}{"id": "abc", "value": 42})

	step := &pipeline.StepConfig{
		ID:   "emitResult",
		Kind: "emit",
		Spec: pipeline.StepSpec{From: "$step1"},
	}

	result, err := EmitNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["id"] != "abc" {
		t.Errorf("expected id=abc, got %v", m["id"])
	}
}

func TestEmitNode_WithExpr(t *testing.T) {
	rctx := newMockCtx(nil)
	rctx.SetOutput("data", map[string]interface{}{
		"name": "test", "count": float64(5),
	})

	step := &pipeline.StepConfig{
		ID:   "emitResult",
		Kind: "emit",
		Spec: pipeline.StepSpec{
			From: "$data",
			Expr: "{name, doubled: (.count * 2)}",
		},
	}

	result, err := EmitNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.(map[string]interface{})
	if m["name"] != "test" {
		t.Errorf("expected name=test, got %v", m["name"])
	}
	if m["doubled"].(float64) != 10 {
		t.Errorf("expected doubled=10, got %v", m["doubled"])
	}
}

func TestEmitNode_WithVars(t *testing.T) {
	rctx := newMockCtx(map[string]interface{}{})
	rctx.SetOutput("itemData", map[string]interface{}{"a": 1, "b": 2})

	step := &pipeline.StepConfig{
		ID:   "emitResult",
		Kind: "emit",
		Spec: pipeline.StepSpec{
			From: "$itemData",
			Vars: map[string]interface{}{"extra": "$itemData"},
			Expr: ". + {sum: ($extra.a + $extra.b)}",
		},
	}

	result, err := EmitNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.(map[string]interface{})
	sum := m["sum"].(int)
	if sum != 3 {
		t.Errorf("expected sum=3, got %v", m["sum"])
	}
}

func TestEmitNode_FromLiteral(t *testing.T) {
	rctx := newMockCtx(nil)

	step := &pipeline.StepConfig{
		ID:   "emitResult",
		Kind: "emit",
		Spec: pipeline.StepSpec{
			From: `{"static": true}`,
		},
	}

	result, err := EmitNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.(map[string]interface{})
	if m["static"] != true {
		t.Errorf("expected static=true, got %v", m["static"])
	}
}

func TestEmitNode_JSONLiteralWithExpr(t *testing.T) {
	rctx := newMockCtx(nil)

	step := &pipeline.StepConfig{
		ID:   "emitResult",
		Kind: "emit",
		Spec: pipeline.StepSpec{
			From: `[1,2,3,4]`,
			Expr: "[.[] | select(. > 2)]",
		},
	}

	result, err := EmitNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := result.([]interface{})
	if len(arr) != 2 {
		t.Errorf("expected 2 items, got %d", len(arr))
	}
}

func TestEmitNode_NoFromNoExpr(t *testing.T) {
	rctx := newMockCtx(nil)

	step := &pipeline.StepConfig{
		ID:   "emitResult",
		Kind: "emit",
		Spec: pipeline.StepSpec{},
	}

	result, err := EmitNode(step, rctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestEmitNode_BadExpr(t *testing.T) {
	rctx := newMockCtx(nil)

	step := &pipeline.StepConfig{
		ID:   "emitResult",
		Kind: "emit",
		Spec: pipeline.StepSpec{
			From: `{}`,
			Expr: "!!!invalid",
		},
	}

	_, err := EmitNode(step, rctx)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestEmitNode_BadRef(t *testing.T) {
	rctx := newMockCtx(nil)

	step := &pipeline.StepConfig{
		ID:   "emitResult",
		Kind: "emit",
		Spec: pipeline.StepSpec{From: "$nothing"},
	}

	_, err := EmitNode(step, rctx)
	if err == nil {
		t.Fatal("expected reference error")
	}
}
