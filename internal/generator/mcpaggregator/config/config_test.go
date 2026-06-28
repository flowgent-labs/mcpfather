package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
aggregateTools:
  - name: test_tool
    description: A test aggregated tool
    inputSchema:
      type: object
      properties:
        id:
          type: string
    pipeline:
      - id: step1
        kind: call
        spec:
          tool: native_tool
          args:
            id: $input.id
      - id: done
        kind: return
        spec:
          from: $step1
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if len(cfg.AggregateTools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(cfg.AggregateTools))
	}
	tool := cfg.AggregateTools[0]
	if tool.Name != "test_tool" {
		t.Errorf("name = %q, want %q", tool.Name, "test_tool")
	}
	if len(tool.Pipeline) != 2 {
		t.Errorf("expected 2 pipeline steps, got %d", len(tool.Pipeline))
	}
}

func TestLoadConfig_Empty(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("LoadConfig should not error for missing file: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.AggregateTools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(cfg.AggregateTools))
	}
}

func TestLoadConfig_WithAnnotations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
aggregateTools:
  - name: mytool
    description: "Test"
    annotations:
      readOnlyHint: true
      destructiveHint: false
    inputSchema:
      type: object
      properties:
        id:
          type: string
    pipeline:
      - id: step1
        kind: return
        spec:
          from: $input.id
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AggregateTools[0].Annotations["readOnlyHint"] != true {
		t.Errorf("expected readOnlyHint=true, got %v", cfg.AggregateTools[0].Annotations["readOnlyHint"])
	}
}
