package converter

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestGetOperations_AllMethods(t *testing.T) {
	op := &openapi3.Operation{Summary: "test"}
	pathItem := &openapi3.PathItem{
		Get:     op,
		Post:    op,
		Put:     op,
		Delete:  op,
		Options: op,
		Head:    op,
		Patch:   op,
		Trace:   op,
	}
	ops := getOperations(pathItem)
	expected := []string{"get", "post", "put", "delete", "options", "head", "patch", "trace"}
	for _, method := range expected {
		if ops[method] != op {
			t.Errorf("expected %s operation to be set", method)
		}
	}
	if len(ops) != len(expected) {
		t.Errorf("expected %d operations, got %d", len(expected), len(ops))
	}
}

func TestGetOperations_None(t *testing.T) {
	ops := getOperations(&openapi3.PathItem{})
	if len(ops) != 0 {
		t.Errorf("expected 0 operations, got %d", len(ops))
	}
}

func TestConvertOperation_RealData(t *testing.T) {
	parser := NewParser(false)
	if err := parser.Parse([]byte(testSpecOAS30)); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	c := &Converter{parser: parser}

	// Pick a path and method that has response content.
	paths := parser.GetPaths()
	if len(paths) == 0 {
		t.Fatal("no paths found in OpenAPI spec")
	}
	var tool *Tool
	for path, pathItem := range paths {
		ops := getOperations(pathItem)
		for method, operation := range ops {
			var err error
			tool, err = c.convertOperation(path, method, operation)
			if err != nil {
				t.Fatalf("convertOperation failed for %s %s: %v", method, path, err)
			}
			if len(tool.Responses) > 0 {
				break
			}
		}
		if tool != nil && len(tool.Responses) > 0 {
			break
		}
	}
	if tool == nil {
		t.Fatal("expected non-nil tool")
	}
	if tool.Name == "" {
		t.Error("expected tool name to be set")
	}
	if !strings.Contains(strings.ToLower(tool.Description), "test") && tool.Description == "" {
		t.Error("expected tool description to be set")
	}
	if tool.RequestTemplate.URL == "" {
		t.Error("expected request template URL to be set")
	}
	if len(tool.Responses) == 0 {
		t.Error("expected at least one response template")
	}
	// Args and RawInputSchema are optional, but you can check them if you want
}
