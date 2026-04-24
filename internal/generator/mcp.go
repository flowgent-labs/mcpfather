package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/lyeslabs/mcpgen/internal/converter"
)

//go:embed templates/*.templ
var templatesFS embed.FS

// ToolTemplateData holds the data to pass to the template for a single tool
type ToolTemplateData struct {
	ToolNameOriginal      string
	ToolNameGo            string
	ToolHandlerName       string
	ToolDescription       string
	RawInputSchema        string
	ResponseTemplate      []converter.ResponseTemplate
	InputSchemaConst      string
	ResponseTemplateConst string
}

// GenerateMCP generates the MCP tool files while preserving existing handler implementations and imports
func (g *Generator) GenerateMCP() error {
	config, err := g.converter.Convert()
	if err != nil {
		return fmt.Errorf("failed at converting OpenAPI schema into MCP code %w", err)
	}

	if err := GenerateGoMod(g.outputDir); err != nil {
		return fmt.Errorf("failed to generate go.mod: %w", err)
	}

	if err := g.GenerateMainGo(); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	if err := g.GenerateServerFile(config); err != nil {
		return fmt.Errorf("failed to generate server file: %w", err)
	}

	if err := g.GenerateToolFiles(config); err != nil {
		return fmt.Errorf("failed to generate tool files: %w", err)
	}

	if err := g.GenerateHelpers(); err != nil {
		return fmt.Errorf("failed to generate helpers: %w", err)
	}

	if err := g.RunGoModTidy(); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %w", err)
	}

	return nil
}

// GenerateMainGo creates the main.go entry point for the standalone project
func (g *Generator) GenerateMainGo() error {
	mainTemplateContent, err := templatesFS.ReadFile("templates/main.templ")
	if err != nil {
		return fmt.Errorf("failed to read main template file: %w", err)
	}

	tmpl, err := template.New("main.templ").Parse(string(mainTemplateContent))
	if err != nil {
		return fmt.Errorf("failed to parse main template: %w", err)
	}

	moduleName := BuildModuleName(g.outputDir)

	data := struct {
		ModuleName string
	}{
		ModuleName: moduleName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to render main template: %w", err)
	}

	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated main.go: %w", err)
	}

	if err := writeFileContent(g.outputDir, "main.go", func() ([]byte, error) {
		return formattedCode, nil
	}); err != nil {
		return fmt.Errorf("failed to write main.go file: %w", err)
	}

	return nil
}

// RunGoModTidy runs `go mod tidy` in the output directory
func (g *Generator) RunGoModTidy() error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = g.outputDir
	cmd.Env = append(os.Environ(), "GOPROXY=https://proxy.golang.org,direct", "GONOSUMCHECK=*", "GOSUMDB=off")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// Rewrite go.mod with module name from output dir
	moduleName := BuildModuleName(g.outputDir)
	goModPath := filepath.Join(g.outputDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod after tidy: %w", err)
	}

	// Replace the module line at the top
	goModContent := string(content)
	if len(goModContent) > 0 {
		newlineIdx := 0
		for i, c := range goModContent {
			if c == '\n' {
				newlineIdx = i
				break
			}
		}
		goModContent = "module " + moduleName + "\n" + goModContent[newlineIdx+1:]
	}

	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to update go.mod module name: %w", err)
	}

	return nil
}
