package engine

import (
	"testing"
)

func TestContext_ResolveInput(t *testing.T) {
	ctx := NewContext(map[string]interface{}{
		"name":  "test",
		"count": 42,
	})
	v, err := ctx.Resolve("$input.name")
	if err != nil {
		t.Fatal(err)
	}
	if v != "test" {
		t.Errorf("expected 'test', got %v", v)
	}
}

func TestContext_ResolvePlainString(t *testing.T) {
	ctx := NewContext(map[string]interface{}{})
	v, err := ctx.Resolve("hello world")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello world" {
		t.Errorf("expected 'hello world', got %v", v)
	}
}

func TestContext_ResolveStepOutput(t *testing.T) {
	ctx := NewContext(map[string]interface{}{})
	ctx.SetOutput("step1", map[string]interface{}{"id": "123"})
	v, err := ctx.Resolve("$step1")
	if err != nil {
		t.Fatal(err)
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", v)
	}
	if m["id"] != "123" {
		t.Errorf("expected id='123', got %v", m["id"])
	}
}

func TestContext_ResolveStepOutputNested(t *testing.T) {
	ctx := NewContext(map[string]interface{}{})
	ctx.SetOutput("policy", map[string]interface{}{"application": map[string]interface{}{"id": "app-456"}})
	v, err := ctx.Resolve("$policy.application.id")
	if err != nil {
		t.Fatal(err)
	}
	if v != "app-456" {
		t.Errorf("expected 'app-456', got %v", v)
	}
}

func TestContext_ResolveItem(t *testing.T) {
	ctx := NewContext(map[string]interface{}{})
	itemCtx := ctx.WithItem(map[string]interface{}{"name": "item1"}, "component")
	v, err := itemCtx.Resolve("$component.name")
	if err != nil {
		t.Fatal(err)
	}
	if v != "item1" {
		t.Errorf("expected 'item1', got %v", v)
	}
}

func TestContext_ResolveItemFull(t *testing.T) {
	ctx := NewContext(map[string]interface{}{})
	itemCtx := ctx.WithItem(map[string]interface{}{"name": "item1", "count": 5}, "component")
	v, err := itemCtx.Resolve("$component")
	if err != nil {
		t.Fatal(err)
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", v)
	}
	if m["name"] != "item1" {
		t.Errorf("expected 'item1', got %v", m["name"])
	}
}

func TestContext_ResolveMap(t *testing.T) {
	ctx := NewContext(map[string]interface{}{"app": "myapp"})
	result, err := ctx.ResolveMap(map[string]interface{}{
		"id":  "$input.app",
		"val": 42,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "myapp" {
		t.Errorf("expected 'myapp', got %v", result["id"])
	}
	if result["val"] != 42 {
		t.Errorf("expected 42, got %v", result["val"])
	}
}

func TestContext_ResolveNested(t *testing.T) {
	ctx := NewContext(map[string]interface{}{
		"user": map[string]interface{}{"profile": map[string]interface{}{"email": "test@test.com"}},
	})
	v, err := ctx.Resolve("$input.user.profile.email")
	if err != nil {
		t.Fatal(err)
	}
	if v != "test@test.com" {
		t.Errorf("expected 'test@test.com', got %v", v)
	}
}

func TestContext_ResolveUnresolved(t *testing.T) {
	ctx := NewContext(map[string]interface{}{})
	_, err := ctx.Resolve("$unknown")
	if err == nil {
		t.Fatal("expected error for unresolved reference")
	}
}

func TestContext_ResolvePathPlain(t *testing.T) {
	ctx := NewContext(map[string]interface{}{})
	v, err := ctx.ResolvePath("literal-string")
	if err != nil {
		t.Fatal(err)
	}
	if v != "literal-string" {
		t.Errorf("expected 'literal-string', got %v", v)
	}
}

func TestContext_ArrayIndexNavigation(t *testing.T) {
	ctx := NewContext(map[string]interface{}{})
	ctx.SetOutput("history", map[string]interface{}{
		"reports": []interface{}{
			map[string]interface{}{"stage": "build"},
			map[string]interface{}{"stage": "test"},
		},
	})
	v, err := ctx.Resolve("$history.reports.0.stage")
	if err != nil {
		t.Fatal(err)
	}
	if v != "build" {
		t.Errorf("expected 'build', got %v", v)
	}
}
