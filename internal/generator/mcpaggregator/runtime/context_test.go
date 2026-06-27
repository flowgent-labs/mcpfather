package runtime

import (
	"testing"
)

func TestContext_ResolveInput(t *testing.T) {
	ctx := NewContext(map[string]interface{}{
		"name": "test",
		"count": 42,
	})
	v, err := ctx.Resolve("{{ input.name }}")
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
	v, err := ctx.Resolve("{{ step1.output.id }}")
	if err != nil {
		t.Fatal(err)
	}
	if v != "123" {
		t.Errorf("expected '123', got %v", v)
	}
}

func TestContext_ResolveItem(t *testing.T) {
	ctx := NewContext(map[string]interface{}{})
	itemCtx := ctx.WithItem(map[string]interface{}{"name": "item1"})
	v, err := itemCtx.Resolve("{{ item.name }}")
	if err != nil {
		t.Fatal(err)
	}
	if v != "item1" {
		t.Errorf("expected 'item1', got %v", v)
	}
}

func TestContext_ResolveMap(t *testing.T) {
	ctx := NewContext(map[string]interface{}{"app": "myapp"})
	result, err := ctx.ResolveMap(map[string]interface{}{
		"id":  "{{ input.app }}",
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
	v, err := ctx.Resolve("{{ input.user.profile.email }}")
	if err != nil {
		t.Fatal(err)
	}
	if v != "test@test.com" {
		t.Errorf("expected 'test@test.com', got %v", v)
	}
}
