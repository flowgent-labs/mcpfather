package converter

import (
	"context"
	"regexp"
	"strings"
	"testing"
)

// testSpecOAS31 is a minimal OAS 3.1 spec with features that the preprocessor
// must convert or strip so kin-openapi (OAS 3.0) can parse them.
const testSpecOAS31 = `openapi: 3.1.0
info:
  title: Blogs API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /posts:
    get:
      operationId: listPosts
      summary: List all blog posts
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            exclusiveMinimum: 0
            exclusiveMaximum: 201
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  results:
                    type: array
                    items:
                      $ref: '#/components/schemas/Post'
    post:
      operationId: createPost
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PostCreate'
      responses:
        "201":
          description: OK
  /posts/{id}:
    get:
      operationId: getPost
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            exclusiveMinimum: 0
      responses:
        "200":
          description: OK
    delete:
      operationId: deletePost
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
            exclusiveMinimum: 0
      responses:
        "204":
          description: OK
  /attachments:
    post:
      operationId: uploadAttachment
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                file:
                  type: string
                  format: binary
      responses:
        "200":
          description: OK
components:
  schemas:
    Post:
      type: object
      properties:
        id:
          type: integer
          exclusiveMinimum: 0
        title:
          type: string
        tags:
          type: array
          prefixItems:
            - type: string
            - type: string
        metadata:
          type: ["object", "null"]
          unevaluatedProperties: false
        status:
          type: string
          const: published
        category:
          type: ["string", "null"]
          if:
            properties:
              type:
                const: featured
          then:
            required: [title]
    PostList:
      type: object
      properties:
        results:
          type: array
          items:
            $ref: '#/components/schemas/Post'
          minContains: 0
          maxContains: 100
          contains:
            type: object
            properties:
              id:
                type: integer
    PostCreate:
      type: object
      required: [title]
      properties:
        title:
          type: string
        body:
          type: ["string", "null"]
`

// TestOAS31Preprocessor verifies that OAS 3.1 features are correctly converted
// or removed so kin-openapi (OAS 3.0) can parse them.
func TestOAS31Preprocessor(t *testing.T) {
	normalized, err := preprocessSpec([]byte(testSpecOAS31))
	if err != nil {
		t.Fatalf("preprocessSpec error: %v", err)
	}

	// Strip YAML comments — comments are preserved by yaml.v3.Marshal and
	// can contain keyword-like text (e.g. "const" in a summary).
	outStr := stripYAMLComments(string(normalized))
	t.Logf("Normalized spec:\n%s", outStr)

	// ─── Conversions ───
	t.Run("exclusiveMinimum", func(t *testing.T) {
		if !strings.Contains(outStr, "exclusiveMinimum: true") {
			t.Error("exclusiveMinimum not converted to bool")
		}
		if !strings.Contains(outStr, "minimum:") {
			t.Error("minimum not inserted alongside exclusiveMinimum")
		}
	})

	t.Run("exclusiveMaximum", func(t *testing.T) {
		if !strings.Contains(outStr, "exclusiveMaximum: true") {
			t.Error("exclusiveMaximum not converted to bool")
		}
		if !strings.Contains(outStr, "maximum:") {
			t.Error("maximum not inserted alongside exclusiveMaximum")
		}
	})

	t.Run("nullable type conversion", func(t *testing.T) {
		if !strings.Contains(outStr, "nullable: true") {
			t.Error("type array with null not converted to nullable")
		}
		for _, pattern := range []string{
			`["string", "null"]`,
			`["boolean", "null"]`,
			`["array", "null"]`,
		} {
			if strings.Contains(outStr, pattern) {
				t.Errorf("type array %s still present", pattern)
			}
		}
	})

	t.Run("prefixItems conversion", func(t *testing.T) {
		if strings.Contains(outStr, "prefixItems:") {
			t.Error("prefixItems still present (should be converted to items)")
		}
	})

	// ─── Removals (checked as YAML key lines to avoid false positives from comments) ───
	removed := map[string]string{
		"const": "JSON Schema constant (const)",
		"$schema": "JSON Schema $schema",
		"$defs": "JSON Schema $defs",
		"prefixItems": "JSON Schema prefixItems",
		"if": "JSON Schema conditional (if)",
		"then": "JSON Schema conditional (then)",
		"contains": "JSON Schema contains",
		"minContains": "JSON Schema minContains",
		"maxContains": "JSON Schema maxContains",
		"unevaluatedProperties": "unevaluatedProperties",
		"unevaluatedItems": "unevaluatedItems",
		"dependentSchemas": "dependentSchemas",
		"contentEncoding": "contentEncoding",
		"contentMediaType": "contentMediaType",
		"examples": "JSON Schema examples array",
		"$dynamicAnchor": "JSON Schema $dynamicAnchor",
		"$dynamicRef": "JSON Schema $dynamicRef",
		"$anchor": "JSON Schema $anchor",
		"jsonSchemaDialect": "OAS 3.1 jsonSchemaDialect",
	}

	reKey := regexp.MustCompile(`(?m)^(\s*)([\w$]+):\s`)
	for keyword, desc := range removed {
		matches := reKey.FindAllStringSubmatch(outStr, -1)
		for _, m := range matches {
			if m[2] == keyword {
				t.Errorf("%s still present as YAML key", desc)
			}
		}
	}

	// ─── kin-openapi parse + validate ───
	t.Run("kin-openapi parse and validate", func(t *testing.T) {
		p := NewParser(true)
		if err := p.Parse(normalized); err != nil {
			t.Fatalf("kin-openapi parse error: %v", err)
		}

		doc := p.GetDocument()
		if err := doc.Validate(context.Background()); err != nil {
			t.Fatalf("kin-openapi validation error: %v", err)
		}

		if p.GetInfo().Title != "Blogs API" {
			t.Errorf("unexpected title: %s", p.GetInfo().Title)
		}

		if len(p.GetPaths()) != 3 {
			t.Errorf("expected 3 paths, got %d", len(p.GetPaths()))
		}
	})
}

// TestOASCompatibility verifies that both OAS 3.0 and 3.1 specs parse
// and validate correctly through the converter.
func TestOASCompatibility(t *testing.T) {
	specs := []struct {
		name          string
		data          []byte
		expectedTitle string
		expectedPaths int
	}{
		{
			name:          "OAS 3.0 Blogs",
			data:          []byte(testSpecOAS30),
			expectedTitle: "Blogs API",
			expectedPaths: 5,
		},
		{
			name:          "OAS 3.1 Blogs",
			data:          []byte(testSpecOAS31),
			expectedTitle: "Blogs API",
			expectedPaths: 3,
		},
	}

	for _, sp := range specs {
		t.Run(sp.name, func(t *testing.T) {
			p := NewParser(true)
			if err := p.Parse(sp.data); err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			doc := p.GetDocument()
			if err := doc.Validate(context.Background()); err != nil {
				t.Fatalf("validation error: %v", err)
			}

			if p.GetInfo().Title != sp.expectedTitle {
				t.Errorf("title = %q, want %q", p.GetInfo().Title, sp.expectedTitle)
			}

			if len(p.GetPaths()) != sp.expectedPaths {
				t.Errorf("paths = %d, want %d", len(p.GetPaths()), sp.expectedPaths)
			}
		})
	}
}

func stripYAMLComments(s string) string {
	var sb strings.Builder
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if idx := strings.Index(line, " #"); idx >= 0 {
			line = line[:idx]
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}
