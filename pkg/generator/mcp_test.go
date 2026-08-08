package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flowgent-labs/mcpfather/pkg/converter"
)

// testConverter implements a minimal converter.Converter interface for testing.
type testConverter struct {
	config *converter.MCPConfig
}

func (tc *testConverter) Convert() (*converter.MCPConfig, error) {
	return tc.config, nil
}

func TestGenerateMCP(t *testing.T) {
	tmpDir := t.TempDir()

	// Prepare a minimal MCPConfig with one tool
	config := &converter.MCPConfig{
		Tools: []converter.Tool{
			{
				Name:           "echo",
				Description:    "Echoes input",
				RawInputSchema: `{"type":"object","properties":{"msg":{"type":"string"}}}`,
				Responses: []converter.ResponseTemplate{
					{PrependBody: "// response", StatusCode: 200, ContentType: "application/json", Suffix: "default"},
				},
				RequestTemplate: converter.RequestTemplate{
					URL:    "/echo",
					Method: "POST",
				},
			},
		},
	}

	// Use the test converter
	g := &Generator{
		PackageName: "mytools",
		outputDir:   tmpDir,
		converter:   &testConverter{config: config},
	}

	// Call GenerateMCP
	if err := g.GenerateMCP(); err != nil {
		t.Fatalf("GenerateMCP failed: %v", err)
	}

	// Check that server.go exists in mcpserver/
	serverGoPath := filepath.Join(tmpDir, "pkg", "mcpserver", "server.go")
	if _, err := os.Stat(serverGoPath); err != nil {
		t.Errorf("Expected mcpserver/server.go to be generated, but it does not exist")
	}

	// Check that client.go exists in mcpserver/helpers/
	helpersGoPath := filepath.Join(tmpDir, "pkg", "helpers", "client.go")
	if _, err := os.Stat(helpersGoPath); err != nil {
		t.Errorf("Expected mcpserver/helpers/client.go to be generated, but it does not exist")
	}

	// Check that the tool file exists in mcpserver/mcptools/
	toolFilePath := filepath.Join(tmpDir, "pkg", "mcptools", "Echo.go")
	data, err := os.ReadFile(toolFilePath)
	if err != nil {
		t.Fatalf("Failed to read generated tool file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "EchoHandler") {
		t.Errorf("Generated tool file missing handler name")
	}
	if !strings.Contains(content, "Echoes input") {
		t.Errorf("Generated tool file missing tool description")
	}
	if !strings.Contains(content, "package mcptools") {
		t.Errorf("Generated tool file missing package declaration")
	}

	mainData, err := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	if err != nil {
		t.Fatalf("Failed to read generated main.go: %v", err)
	}
	mainContent := string(mainData)
	for _, want := range []string{
		`version := flag.Bool("version", false, "Print version and exit")`,
		`flag.BoolVar(version, "V", false, "Print version and exit")`,
		`func printVersion()`,
		`runtime.Version()`,
	} {
		if !strings.Contains(mainContent, want) {
			t.Errorf("generated main.go missing %q", want)
		}
	}

	makefileData, err := os.ReadFile(filepath.Join(tmpDir, "Makefile"))
	if err != nil {
		t.Fatalf("Failed to read generated Makefile: %v", err)
	}
	makefileContent := string(makefileData)
	for _, want := range []string{
		`LDFLAGS := -ldflags`,
		`-X main.versionStr=$(VERSION)`,
		`-X main.gitCommit=$(GIT_COMMIT)`,
		`-X main.buildDate=$(BUILD_DATE)`,
		`--build-arg VERSION=$(VERSION)`,
	} {
		if !strings.Contains(makefileContent, want) {
			t.Errorf("generated Makefile missing %q", want)
		}
	}

	dockerfileData, err := os.ReadFile(filepath.Join(tmpDir, "deploy", "docker", "Dockerfile"))
	if err != nil {
		t.Fatalf("Failed to read generated Dockerfile: %v", err)
	}
	dockerfileContent := string(dockerfileData)
	for _, want := range []string{
		`ARG VERSION=dev`,
		`ARG GIT_COMMIT=unknown`,
		`ARG BUILD_DATE=unknown`,
		`-X main.versionStr=${VERSION}`,
	} {
		if !strings.Contains(dockerfileContent, want) {
			t.Errorf("generated Dockerfile missing %q", want)
		}
	}

	licenseData, err := os.ReadFile(filepath.Join(tmpDir, "LICENSE"))
	if err != nil {
		t.Fatalf("Failed to read generated LICENSE: %v", err)
	}
	licenseContent := string(licenseData)
	for _, want := range []string{
		`Apache License`,
		`Version 2.0, January 2004`,
		`TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION`,
	} {
		if !strings.Contains(licenseContent, want) {
			t.Errorf("generated LICENSE missing %q", want)
		}
	}
	if strings.Contains(licenseContent, "MIT License") || strings.Contains(licenseContent, "GNU AFFERO GENERAL PUBLIC LICENSE") {
		t.Errorf("generated LICENSE should be Apache-2.0 only")
	}

	readmeData, err := os.ReadFile(filepath.Join(tmpDir, "README.md"))
	if err != nil {
		t.Fatalf("Failed to read generated README.md: %v", err)
	}
	readmeContent := string(readmeData)
	for _, want := range []string{
		`Apache License 2.0`,
		`commercial use`,
		`not subject to mcpfather's AGPL`,
	} {
		if !strings.Contains(readmeContent, want) {
			t.Errorf("generated README.md missing license statement %q", want)
		}
	}
}

// TestBacktickInMarkdown verifies that backticks in markdown descriptions
// don't break Go raw string literal generation.
func TestBacktickInMarkdown(t *testing.T) {
	tmpDir := t.TempDir()

	markdownWithBackticks := `# API Response

The response contains ` + "`" + `code blocks` + "`" + ` and ` + "`" + `inline code` + "`" + `.

Example:
` + "```json" + `
{"key": "value"}
` + "```" + `
`

	config := &converter.MCPConfig{
		Tools: []converter.Tool{
			{
				Name:           "withBackticks",
				Description:    "Tool with backticks in description",
				RawInputSchema: `{"type":"object"}`,
				Responses: []converter.ResponseTemplate{
					{PrependBody: markdownWithBackticks, StatusCode: 200, ContentType: "application/json", Suffix: "A"},
				},
				RequestTemplate: converter.RequestTemplate{
					URL:    "/test",
					Method: "GET",
				},
			},
		},
	}

	g := &Generator{
		PackageName: "backticktest",
		outputDir:   tmpDir,
		converter:   &testConverter{config: config},
	}

	if err := g.GenerateMCP(); err != nil {
		t.Fatalf("GenerateMCP with backticks failed: %v", err)
	}

	// Verify the generated tool file compiles
	toolFile := filepath.Join(tmpDir, "pkg", "mcptools", "WithBackticks.go")
	data, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("Failed to read generated tool file: %v", err)
	}

	// The backticks should be escaped as \x60 in double-quoted strings
	content := string(data)
	if !strings.Contains(content, `\x60`) {
		t.Error("Backtick escape sequence (\\x60) not found in generated code")
	}

	// Verify the generated code actually compiles
	proxyURL := os.Getenv("MCPFATHER_TEST_PROXY")
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTPS_PROXY")
	}
	buildEnv := append(os.Environ(),
		"GOPROXY=https://goproxy.cn,direct",
		"GOSUMDB=off",
		"GOTOOLCHAIN=local",
	)
	if proxyURL != "" {
		buildEnv = append(buildEnv, "HTTPS_PROXY="+proxyURL)
	}
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	tidyCmd.Env = buildEnv
	if out, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed:\n%s", out)
	}
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = tmpDir
	cmd.Env = buildEnv
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Generated code does not compile:\n%s\n%s", out, content)
	}
}
