package converter

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func getTestSpecBytes(t *testing.T) []byte {
	return []byte(testSpecOAS30)
}

func TestNewParser(t *testing.T) {
	p := NewParser(true)
	if p == nil {
		t.Fatal("expected non-nil parser")
	}
	if !p.ValidateDocument {
		t.Error("expected ValidateDocument to be true")
	}
}

func TestParser_Parse_ValidYAML(t *testing.T) {
	p := NewParser(false)
	data := getTestSpecBytes(t)
	err := p.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	doc := p.GetDocument()
	if doc == nil {
		t.Fatal("expected parsed document")
	}
	if doc.Info == nil || doc.Info.Title == "" {
		t.Errorf("expected non-empty info title")
	}
}

func TestParser_ParseFile(t *testing.T) {
	p := NewParser(false)
	specPath := writeTestSpecOAS30(t)
	err := p.ParseFile(specPath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	doc := p.GetDocument()
	if doc == nil {
		t.Fatal("expected parsed document from file")
	}
}

func TestParser_ParseXquikOAS31Fixture(t *testing.T) {
	p := NewParser(true)
	err := p.ParseFile("../../e2e/testdata/xquik_spec.yaml")
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	paths := p.GetPaths()
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(paths))
	}

	search, ok := paths["/api/v1/x/tweets/search"]
	if !ok {
		t.Fatal("expected search path")
	}
	if search.Get == nil || search.Get.OperationID != "searchTweets" {
		t.Fatalf("unexpected search operation: %#v", search.Get)
	}

	user, ok := paths["/api/v1/x/users/{id}"]
	if !ok {
		t.Fatal("expected user path")
	}
	if user.Get == nil || user.Get.OperationID != "getUser" {
		t.Fatalf("unexpected user operation: %#v", user.Get)
	}
}

func TestParser_GetDocument(t *testing.T) {
	p := NewParser(false)
	_ = p.Parse(getTestSpecBytes(t))
	doc := p.GetDocument()
	if doc == nil {
		t.Fatal("expected non-nil document")
	}
}

func TestParser_GetPaths(t *testing.T) {
	p := NewParser(false)
	_ = p.Parse(getTestSpecBytes(t))
	paths := p.GetPaths()
	if paths == nil {
		t.Fatal("expected non-nil paths")
	}
	// Optionally check for a known path in your example_confluence_oas_v3.0.yaml
	// if _, ok := paths["/hello"]; !ok {
	// 	t.Errorf("expected /hello path")
	// }
}

func TestParser_GetServers(t *testing.T) {
	p := NewParser(false)
	_ = p.Parse(getTestSpecBytes(t))
	servers := p.GetServers()
	if len(servers) == 0 {
		t.Fatal("expected at least one server")
	}
}

func TestParser_GetInfo(t *testing.T) {
	p := NewParser(false)
	_ = p.Parse(getTestSpecBytes(t))
	info := p.GetInfo()
	if info == nil {
		t.Fatal("expected non-nil info")
	}
}

func TestParser_GetOperationID_WithExplicitID(t *testing.T) {
	p := NewParser(false)
	op := &openapi3.Operation{OperationID: "explicitID"}
	id := p.GetOperationID("/foo", "GET", op)
	if id != "explicitID" {
		t.Errorf("expected explicitID, got %q", id)
	}
}

func TestParser_GetOperationID_Generated(t *testing.T) {
	p := NewParser(false)
	op := &openapi3.Operation{}
	id := p.GetOperationID("/foo/{bar}", "POST", op)
	if !strings.HasPrefix(id, "post_foo_bar") {
		t.Errorf("expected generated id to start with post_foo_bar, got %q", id)
	}
}
