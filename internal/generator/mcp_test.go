package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lyeslabs/mcpgen/internal/converter"
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
	serverGoPath := filepath.Join(tmpDir, "internal", "mcpserver", "server.go")
	if _, err := os.Stat(serverGoPath); err != nil {
		t.Errorf("Expected mcpserver/server.go to be generated, but it does not exist")
	}

	// Check that client.go exists in mcpserver/helpers/
	helpersGoPath := filepath.Join(tmpDir, "internal", "helpers", "client.go")
	if _, err := os.Stat(helpersGoPath); err != nil {
		t.Errorf("Expected mcpserver/helpers/client.go to be generated, but it does not exist")
	}

	// Check that the tool file exists in mcpserver/mcptools/
	toolFilePath := filepath.Join(tmpDir, "internal", "mcptools", "Echo.go")
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
	toolFile := filepath.Join(tmpDir, "internal", "mcptools", "WithBackticks.go")
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
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	tidyCmd.Env = append(os.Environ(), "GOPROXY=https://proxy.golang.org,direct", "GONOSUMCHECK=*", "GOSUMDB=off")
	if out, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed:\n%s", out)
	}
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = tmpDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Generated code does not compile:\n%s\n%s", out, content)
	}
}
