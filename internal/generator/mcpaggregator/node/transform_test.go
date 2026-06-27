package node

import (
	"reflect"
	"testing"
)

func TestApplyProject(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"age":  42,
		"city": "NYC",
	}
	result := applyProject(data, []string{"name", "city"})
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}
	if len(m) != 2 {
		t.Errorf("expected 2 fields, got %d", len(m))
	}
	if m["name"] != "test" || m["city"] != "NYC" {
		t.Errorf("unexpected project result: %v", m)
	}
	if _, ok := m["age"]; ok {
		t.Error("age should have been excluded")
	}
}

func TestApplyRemove(t *testing.T) {
	data := map[string]interface{}{
		"name":    "test",
		"private": "secret",
	}
	result := applyRemove(data, []string{"private"})
	m := result.(map[string]interface{})
	if _, ok := m["private"]; ok {
		t.Error("private should have been removed")
	}
}

func TestApplyRename(t *testing.T) {
	data := map[string]interface{}{
		"old_name": "value",
	}
	result := applyRename(data, map[string]string{"old_name": "new_name"})
	m := result.(map[string]interface{})
	if m["new_name"] != "value" {
		t.Errorf("expected 'value', got %v", m["new_name"])
	}
	if _, ok := m["old_name"]; ok {
		t.Error("old_name should be gone")
	}
}

func TestApplyDefault(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
	}
	result := applyDefault(data, map[string]interface{}{
		"name":  "default",
		"count": 1,
	})
	m := result.(map[string]interface{})
	if m["name"] != "test" {
		t.Error("existing value should not be overwritten")
	}
	if m["count"] != 1 {
		t.Error("missing field should get default")
	}
}

func TestApplyFlatten(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"nested": map[string]interface{}{
			"inner": "value",
		},
	}
	result := applyFlatten(data, []string{"nested"})
	m := result.(map[string]interface{})
	if m["inner"] != "value" {
		t.Errorf("expected 'value', got %v", m["inner"])
	}
	if _, ok := m["nested"]; ok {
		t.Error("nested should be gone")
	}
}

func TestDeepCopy(t *testing.T) {
	orig := map[string]interface{}{
		"list": []interface{}{1, 2, 3},
		"nested": map[string]interface{}{"key": "value"},
	}
	cp := deepCopy(orig)
	if !reflect.DeepEqual(orig, cp) {
		t.Error("copy should equal original")
	}
	cpMap := cp.(map[string]interface{})
	cpMap["nested"].(map[string]interface{})["key"] = "modified"
	if orig["nested"].(map[string]interface{})["key"] != "value" {
		t.Error("deep copy should not modify original")
	}
}
