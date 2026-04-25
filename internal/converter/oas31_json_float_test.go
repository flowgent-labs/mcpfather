package converter

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestOAS31FloatExclusives(t *testing.T) {
	data := []byte(`openapi: 3.1.0
info:
  title: Float Exclusive Test
  version: 1.0.0
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  value:
                    type: number
                    exclusiveMinimum: 0.0
                    exclusiveMaximum: 100.5
`)
	normalized, err := preprocessSpec(data)
	if err != nil {
		t.Fatalf("preprocessSpec error: %v", err)
	}

	outStr := string(normalized)
	t.Logf("Normalized:\n%s", outStr)

	if !strings.Contains(outStr, "exclusiveMinimum: true") {
		t.Error("exclusiveMinimum not converted")
	}
	if !strings.Contains(outStr, "minimum: 0.0") {
		t.Error("minimum not preserved as float 0.0")
	}
	if !strings.Contains(outStr, "maximum: 100.5") {
		t.Error("maximum not preserved as float 100.5")
	}

	p := NewParser(true)
	if err := p.Parse(normalized); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	doc := p.GetDocument()
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("validation error: %v", err)
	}
}

func TestOAS31ExclusivesWithExistingMinMax(t *testing.T) {
	// Simulates real-world specs where both exclusiveMinimum and minimum coexist
	// (common in SpringDoc/Swagger-generated specs)
	data := []byte(`openapi: 3.1.0
info:
  title: Existing Min Max Test
  version: 1.0.0
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  count:
                    type: integer
                    exclusiveMinimum: 0.0
                    minimum: 0
                    exclusiveMaximum: 100.0
                    maximum: 100
`)
	normalized, err := preprocessSpec(data)
	if err != nil {
		t.Fatalf("preprocessSpec error: %v", err)
	}

	outStr := string(normalized)
	t.Logf("Normalized:\n%s", outStr)

	// exclusiveMinimum should become true
	if !strings.Contains(outStr, "exclusiveMinimum: true") {
		t.Error("exclusiveMinimum not converted")
	}
	// minimum: 0 should still exist (original)
	if !strings.Contains(outStr, "minimum:") {
		t.Error("original minimum missing")
	}

	p := NewParser(true)
	if err := p.Parse(normalized); err != nil {
		t.Fatalf("parse error: %v", err)
	}
}

func TestOAS31JSONInput(t *testing.T) {
	jsonData := []byte(`{
  "openapi": "3.1.0",
  "info": {"title": "JSON Input Test", "version": "1.0.0"},
  "paths": {
    "/test": {
      "get": {
        "operationId": "getTest",
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "name": {"type": ["string", "null"]},
                    "count": {
                      "type": "integer",
                      "exclusiveMinimum": 0,
                      "exclusiveMaximum": 100
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`)

	normalized, err := preprocessSpec(jsonData)
	if err != nil {
		t.Fatalf("preprocessSpec error: %v", err)
	}
	t.Logf("Normalized JSON output:\n%s", normalized)

	p := NewParser(true)
	if err := p.Parse(normalized); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	doc := p.GetDocument()
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("validation error: %v", err)
	}
	if p.GetInfo().Title != "JSON Input Test" {
		t.Errorf("title = %q, want %q", p.GetInfo().Title, "JSON Input Test")
	}
}

func TestOAS31JSONInputThroughParser(t *testing.T) {
	jsonData := []byte(`{
  "openapi": "3.1.0",
  "info": {"title": "Parser JSON Test", "version": "1.0.0"},
  "paths": {},
  "components": {
    "schemas": {
      "Foo": {
        "type": ["string", "null"],
        "exclusiveMinimum": 0.0
      }
    }
  }
}`)

	p := NewParser(true)
	if err := p.Parse(jsonData); err != nil {
		t.Fatalf("Parser.Parse JSON error: %v", err)
	}
	if p.GetInfo().Title != "Parser JSON Test" {
		t.Errorf("title = %q, want %q", p.GetInfo().Title, "Parser JSON Test")
	}
}

func TestOAS31ConfluenceJSON(t *testing.T) {
	// Confluence OAS 3.1 spec (YAML) with numeric exclusives and nullable types
	data, err := os.ReadFile("../../testdata/example_confluence_oas_v3.1.yaml")
	if err != nil {
		t.Fatal(err)
	}

	p := NewParser(true)
	if err := p.Parse(data); err != nil {
		t.Fatalf("Parser.Parse Spring Boot JSON error: %v", err)
	}

	doc := p.GetDocument()
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("validation error: %v", err)
	}

	if p.GetInfo().Title != "Confluence Cloud REST API" {
		t.Errorf("title = %q, want %q", p.GetInfo().Title, "Confluence Cloud REST API")
	}
	if len(p.GetPaths()) != 11 {
		t.Errorf("paths = %d, want 11", len(p.GetPaths()))
	}
}

