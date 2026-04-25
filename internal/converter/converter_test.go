package converter

import (
	"os"
	"path/filepath"
	"testing"
)

var specPath = filepath.Join("..", "..", "testdata", "example_confluence_oas_v3.0.yaml")

func TestNewConverter(t *testing.T) {
	parser := NewParser(false)
	c, err := NewConverter(parser, nil, nil)
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

	c, err := NewConverter(parser, nil, nil)
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

func TestPathMatch_TrailingSlashes(t *testing.T) {
	tests := []struct {
		specPath  string
		filter    string
		method    string
		wantMatch bool
	}{
		{"/api/v2/login", "/api/v2/login/", "post", true},
		{"/api/v2/login/", "/api/v2/login", "post", true},
		{"/api/v2/login", "/api/v2/login", "post", true},
		{"/api/v2/scans/{scan_id}", "/api/v2/scans/{id}", "get", true},
		{"/api/v2/scans/{scan_id}", "/api/v2/scans/{scan_id}", "get", true},
		{"/api/v2/scans/{scan_id}", "/api/v2/scans/{other}/details", "get", false},
		{"/api/v2/scans", "/api/v2/scans/{id}", "get", false},
		{"/health", "/api/v2/health", "get", false},
	}
	for _, tt := range tests {
		t.Run(tt.specPath+"/"+tt.filter, func(t *testing.T) {
			got := pathMatch(tt.specPath, tt.filter, tt.method)
			if got != tt.wantMatch {
				t.Errorf("pathMatch(%q, %q, %q) = %v, want %v", tt.specPath, tt.filter, tt.method, got, tt.wantMatch)
			}
		})
	}
}

func TestConverter_Convert_NoDocument(t *testing.T) {
	parser := NewParser(false)
	c, err := NewConverter(parser, nil, nil)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	_, err = c.Convert()
	if err == nil {
		t.Fatal("expected error when no OpenAPI document is loaded")
	}
}
