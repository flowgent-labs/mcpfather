package converter

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestOAS31Preprocessor verifies that OAS 3.1 features are correctly converted
// or removed so kin-openapi (OAS 3.0) can parse them.
func TestOAS31Preprocessor(t *testing.T) {
	data, err := os.ReadFile("../../testdata/example_confluence_oas_v3.1.yaml")
	if err != nil {
		t.Fatal(err)
	}

	normalized, err := preprocessSpec(data)
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

		if p.GetInfo().Title != "Confluence Cloud REST API" {
			t.Errorf("unexpected title: %s", p.GetInfo().Title)
		}

		if len(p.GetPaths()) != 12 {
			t.Errorf("expected 12 paths, got %d", len(p.GetPaths()))
		}
	})
}

// TestOASCompatibility verifies that both OAS 3.0 and 3.1 specs parse
// and validate correctly through the converter.
func TestOASCompatibility(t *testing.T) {
	specs := []struct {
		name, path       string
		expectedTitle    string
		expectedPaths    int
	}{
		{
			name:          "OAS 3.0 Confluence",
			path:          "../../testdata/example_confluence_oas_v3.0.yaml",
			expectedTitle: "Confluence Cloud REST API",
			expectedPaths: 12,
		},
		{
			name:          "OAS 3.1 Confluence",
			path:          "../../testdata/example_confluence_oas_v3.1.yaml",
			expectedTitle: "Confluence Cloud REST API",
			expectedPaths: 12,
		},
	}

	for _, sp := range specs {
		t.Run(sp.name, func(t *testing.T) {
			p := NewParser(true)
			if err := p.ParseFile(sp.path); err != nil {
				t.Fatalf("ParseFile error: %v", err)
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
