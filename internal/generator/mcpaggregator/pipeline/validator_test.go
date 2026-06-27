package pipeline

import (
	"testing"
)

func TestValidate_ValidPipeline(t *testing.T) {
	steps := []StepConfig{
		{
			Name: "step1",
			Type: "call",
			Call: &CallConfig{Tool: "myTool", Args: map[string]interface{}{"id": "{{ input.id }}"}},
		},
		{
			Name: "done",
			Type: "return",
			Return: &ReturnConfig{Source: "step1.output"},
		},
	}
	if err := Validate(steps); err != nil {
		t.Errorf("valid pipeline should not error: %v", err)
	}
}

func TestValidate_DuplicateName(t *testing.T) {
	steps := []StepConfig{
		{Name: "step1", Type: "call", Call: &CallConfig{Tool: "t"}},
		{Name: "step1", Type: "call", Call: &CallConfig{Tool: "t"}},
	}
	if err := Validate(steps); err == nil {
		t.Error("duplicate names should error")
	}
}

func TestValidate_MissingName(t *testing.T) {
	steps := []StepConfig{
		{Type: "call", Call: &CallConfig{Tool: "t"}},
	}
	if err := Validate(steps); err == nil {
		t.Error("missing name should error")
	}
}

func TestValidate_UnknownType(t *testing.T) {
	steps := []StepConfig{
		{Name: "step1", Type: "unknown"},
	}
	if err := Validate(steps); err == nil {
		t.Error("unknown type should error")
	}
}

func TestValidate_MissingCallConfig(t *testing.T) {
	steps := []StepConfig{
		{Name: "step1", Type: "call"},
	}
	if err := Validate(steps); err == nil {
		t.Error("missing call config should error")
	}
}

func TestValidate_MissingReturnSource(t *testing.T) {
	steps := []StepConfig{
		{Name: "done", Type: "return", Return: &ReturnConfig{}},
	}
	if err := Validate(steps); err == nil {
		t.Error("missing return source should error")
	}
}

func TestValidate_DuplicateOutput(t *testing.T) {
	steps := []StepConfig{
		{Name: "step1", Type: "call", Call: &CallConfig{Tool: "t"}, Output: "out"},
		{Name: "step2", Type: "call", Call: &CallConfig{Tool: "t"}, Output: "out"},
	}
	if err := Validate(steps); err == nil {
		t.Error("duplicate outputs should error")
	}
}
