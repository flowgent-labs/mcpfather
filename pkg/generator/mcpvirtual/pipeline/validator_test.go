package pipeline

import (
	"testing"
)

func TestValidate_ValidPipeline(t *testing.T) {
	steps := []StepConfig{
		{ID: "getData", Kind: "call", Spec: StepSpec{Tool: "GetData", Args: map[string]interface{}{"id": "$input.id"}}},
		{ID: "transform", Kind: "jq", Spec: StepSpec{From: "$getData", Expr: ".items"}},
		{ID: "done", Kind: "return", Spec: StepSpec{From: "$transform"}},
	}
	if err := Validate(steps); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_DuplicateID(t *testing.T) {
	steps := []StepConfig{
		{ID: "step1", Kind: "return", Spec: StepSpec{From: "$input"}},
		{ID: "step1", Kind: "return", Spec: StepSpec{From: "$input"}},
	}
	if err := Validate(steps); err == nil {
		t.Fatal("expected error for duplicate id")
	}
}

func TestValidate_MissingID(t *testing.T) {
	steps := []StepConfig{
		{ID: "", Kind: "return", Spec: StepSpec{From: "$input"}},
	}
	if err := Validate(steps); err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestValidate_UnknownKind(t *testing.T) {
	steps := []StepConfig{
		{ID: "step1", Kind: "unknown"},
	}
	if err := Validate(steps); err == nil {
		t.Fatal("expected error for unknown kind")
	}
}

func TestValidate_CallMissingTool(t *testing.T) {
	steps := []StepConfig{
		{ID: "step1", Kind: "call", Spec: StepSpec{Args: map[string]interface{}{}}},
	}
	if err := Validate(steps); err == nil {
		t.Fatal("expected error for call without tool")
	}
}

func TestValidate_CallMissingArgs(t *testing.T) {
	steps := []StepConfig{
		{ID: "step1", Kind: "call", Spec: StepSpec{Tool: "SomeTool"}},
	}
	if err := Validate(steps); err == nil {
		t.Fatal("expected error for call without args")
	}
}

func TestValidate_JQMissingExpr(t *testing.T) {
	steps := []StepConfig{
		{ID: "step1", Kind: "jq", Spec: StepSpec{From: "$data"}},
	}
	if err := Validate(steps); err == nil {
		t.Fatal("expected error for jq without expr")
	}
}

func TestValidate_ForeachMissingIn(t *testing.T) {
	steps := []StepConfig{
		{ID: "step1", Kind: "foreach", Spec: StepSpec{As: "item", Pipeline: []StepConfig{{ID: "emit1", Kind: "emit"}}}},
	}
	if err := Validate(steps); err == nil {
		t.Fatal("expected error for foreach without in")
	}
}

func TestValidate_ForeachMissingAs(t *testing.T) {
	steps := []StepConfig{
		{ID: "step1", Kind: "foreach", Spec: StepSpec{In: "$data", Pipeline: []StepConfig{{ID: "emit1", Kind: "emit"}}}},
	}
	if err := Validate(steps); err == nil {
		t.Fatal("expected error for foreach without as")
	}
}

func TestValidate_ForeachMissingEmit(t *testing.T) {
	steps := []StepConfig{
		{ID: "step1", Kind: "foreach", Spec: StepSpec{In: "$data", As: "item", Pipeline: []StepConfig{
			{ID: "call1", Kind: "call", Spec: StepSpec{Tool: "T", Args: map[string]interface{}{}}},
		}}},
	}
	if err := Validate(steps); err == nil {
		t.Fatal("expected error for foreach without emit")
	}
}

func TestValidate_ForeachNoReturn(t *testing.T) {
	steps := []StepConfig{
		{ID: "step1", Kind: "foreach", Spec: StepSpec{In: "$data", As: "item", Pipeline: []StepConfig{
			{ID: "ret1", Kind: "return", Spec: StepSpec{From: "$item"}},
		}}},
	}
	if err := Validate(steps); err == nil {
		t.Fatal("expected error for return inside foreach")
	}
}

func TestValidate_ForeachValid(t *testing.T) {
	steps := []StepConfig{
		{ID: "step1", Kind: "foreach", Spec: StepSpec{In: "$data", As: "item", Pipeline: []StepConfig{
			{ID: "emit1", Kind: "emit", Spec: StepSpec{From: "$item"}},
		}}},
	}
	if err := Validate(steps); err != nil {
		t.Fatal(err)
	}
}

func TestValidateReferences_Valid(t *testing.T) {
	steps := []StepConfig{
		{ID: "getData", Kind: "call", Spec: StepSpec{Tool: "GetData", Args: map[string]interface{}{"id": "$input.dataId"}}},
		{ID: "transform", Kind: "jq", Spec: StepSpec{From: "$getData", Expr: ".items"}},
		{ID: "done", Kind: "return", Spec: StepSpec{From: "$transform"}},
	}
	if err := ValidateReferences(steps); err != nil {
		t.Fatal(err)
	}
}

func TestValidateReferences_Unresolved(t *testing.T) {
	steps := []StepConfig{
		{ID: "getData", Kind: "call", Spec: StepSpec{Tool: "GetData", Args: map[string]interface{}{"id": "$unknownRef"}}},
	}
	if err := ValidateReferences(steps); err == nil {
		t.Fatal("expected error for unresolved reference")
	}
}

func TestValidateReferences_ForeachItemRef(t *testing.T) {
	steps := []StepConfig{
		{ID: "step1", Kind: "foreach", Spec: StepSpec{In: "$input.items", As: "item", Pipeline: []StepConfig{
			{ID: "emit1", Kind: "emit", Spec: StepSpec{From: "$item"}},
		}}},
	}
	if err := ValidateReferences(steps); err != nil {
		t.Fatal(err)
	}
}
