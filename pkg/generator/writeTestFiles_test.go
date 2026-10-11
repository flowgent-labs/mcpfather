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
		"mcptools_test_helpers.templ",
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

func TestGeneratedMCPToolTestsInstallLoadedConfig(t *testing.T) {
	outputDir := t.TempDir()
	g := &Generator{outputDir: outputDir}
	if err := g.GenerateCommonTestFiles(); err != nil {
		t.Fatalf("GenerateCommonTestFiles: %v", err)
	}

	helperPath := filepath.Join(outputDir, "pkg", "mcptools", "test_helpers_test.go")
	helper, err := os.ReadFile(helperPath)
	if err != nil {
		t.Fatalf("read generated test helper: %v", err)
	}
	helperText := string(helper)
	for _, expected := range []string{
		`t.Setenv("HOME", t.TempDir())`,
		"cfg, err := mcputils.LoadConfig(serviceName)",
		"mcputils.SetConfig(cfg)",
		"t.Cleanup(func()",
		"mcputils.SetConfig(previous)",
	} {
		if !strings.Contains(helperText, expected) {
			t.Errorf("generated test helper is missing %q", expected)
		}
	}

	registryPath := filepath.Join(outputDir, "pkg", "mcptools", "registry_test.go")
	registry, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("read generated registry test: %v", err)
	}
	if got := strings.Count(string(registry), "installTestConfig(t,"); got != 3 {
		t.Fatalf("registry test installs config %d times, want 3", got)
	}

	toolTemplate, err := templatesFS.ReadFile("templates/tool_test.templ")
	if err != nil {
		t.Fatalf("read tool test template: %v", err)
	}
	toolTemplateText := string(toolTemplate)
	if got := strings.Count(toolTemplateText, "installTestConfig(t,"); got != 3 {
		t.Fatalf("tool test template installs config %d times, want 3", got)
	}
	if strings.Contains(toolTemplateText, "mcputils.LoadConfig(") {
		t.Fatal("tool test template discards a loaded config")
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
