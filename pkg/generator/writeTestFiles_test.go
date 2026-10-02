package generator

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommonTemplateNamesMatchGeneratedFiles(t *testing.T) {
	for _, name := range []string{
		"client.templ",
		"client_test.templ",
		"config.templ",
		"config_test.templ",
		"resource_server.templ",
		"resource_server_test.templ",
	} {
		if _, err := templatesFS.ReadFile("templates/" + name); err != nil {
			t.Errorf("read templates/%s: %v", name, err)
		}
	}

	for _, legacyName := range []string{
		"helpers.templ",
		"helpers_client_test.templ",
		"helpers_config_test.templ",
		"helpers_resource_server_test.templ",
		"mcpconfig_types_test.templ",
	} {
		if _, err := templatesFS.ReadFile("templates/" + legacyName); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("legacy template %s still exists or returned an unexpected error: %v", legacyName, err)
		}
	}
}

func TestGenerateConfigCopiesCanonicalTypesTest(t *testing.T) {
	outputDir := t.TempDir()
	g := &Generator{outputDir: outputDir}
	if err := g.GenerateConfig(); err != nil {
		t.Fatalf("GenerateConfig: %v", err)
	}

	generatedPath := filepath.Join(outputDir, "pkg", "mcpconfig", "types_test.go")
	generated, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("read generated types_test.go: %v", err)
	}
	text := string(generated)
	if !strings.Contains(text, "func TestMiddlewareConfig_Fields") {
		t.Fatal("generated types_test.go did not come from the canonical mcpconfig test source")
	}
	if strings.Contains(text, "TestMiddlewareConfigPolicyMaps") {
		t.Fatal("generated types_test.go contains the removed duplicate template test")
	}
}
