package converter

import (
	"os"
	"path/filepath"
	"testing"
)

var specPath = filepath.Join("..", "..", "testdata", "example_confluence_oas_v3.0.yaml")

func TestNewConverter(t *testing.T) {
	parser := NewParser(false)
	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil Converter")
	}
	if c.parser != parser {
		t.Error("expected parser to be set")
	}
	if c.options.ServerConfig == nil {
		t.Error("expected ServerConfig to be initialized")
	}
}

func TestConverter_Convert(t *testing.T) {
	// Load a real OpenAPI spec
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Fatalf("Test setup error: fixture file %s does not exist. Please create it.", specPath)
	}
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("could not read %s: %v", specPath, err)
	}
	parser := NewParser(false)
	if err := parser.Parse(data); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil MCPConfig")
	}
	if config.Server.Config == nil {
		t.Error("expected Server.Config to be set")
	}
	if len(config.Tools) == 0 {
		t.Error("expected at least one tool in Tools")
	}
	// Check that tools are sorted by name
	for i := 1; i < len(config.Tools); i++ {
		if config.Tools[i-1].Name > config.Tools[i].Name {
			t.Errorf("tools not sorted by name: %q > %q", config.Tools[i-1].Name, config.Tools[i].Name)
		}
	}
}

func TestCleanOperationId(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"listSpaces", "listSpaces"},
		{"'listSpaces'", "listSpaces"},
		{`"listSpaces"`, "listSpaces"},
		{"  listSpaces  ", "listSpaces"},
		{"listSpaces\n", "listSpaces"},
		{"listSpaces\r\n", "listSpaces"},
		{"get-a-very-long-operation-id", "get-a-very-long-operation-id"},
		{"", ""},
		{"''", ""},
		{`""`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanOperationId(tt.input)
			if got != tt.want {
				t.Errorf("cleanOperationId(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConverter_Convert_IncludeExcludeByOperationId(t *testing.T) {
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Fatalf("fixture file %s does not exist", specPath)
	}
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("could not read %s: %v", specPath, err)
	}

	parser := NewParser(false)
	if err := parser.Parse(data); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	// Include only "listSpaces"
	c, err := NewConverter(parser, []string{"listSpaces"}, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if len(config.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(config.Tools))
	}
	if config.Tools[0].Name != "Listspaces" {
		t.Errorf("expected tool Listspaces, got %s", config.Tools[0].Name)
	}
}

func TestConverter_Convert_NoDocument(t *testing.T) {
	parser := NewParser(false)
	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	_, err = c.Convert()
	if err == nil {
		t.Fatal("expected error when no OpenAPI document is loaded")
	}
}
