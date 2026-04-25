package generator

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/lyeslabs/mcpgen/internal/converter"
)

//go:embed templates/*.templ
//go:embed templates/_credentials/*
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

	if err := g.GenerateCredentials(); err != nil {
		return fmt.Errorf("failed to generate credentials: %w", err)
	}

	if err := g.GenerateClientSh(config); err != nil {
		return fmt.Errorf("failed to generate client script: %w", err)
	}

	if err := g.GenerateMakefile(); err != nil {
		return fmt.Errorf("failed to generate Makefile: %w", err)
	}

	if err := g.GenerateReadme(); err != nil {
		return fmt.Errorf("failed to generate README.md: %w", err)
	}

	if err := g.GenerateDotCredentials(); err != nil {
		return fmt.Errorf("failed to generate .credentials: %w", err)
	}

	if err := g.GenerateDotGitignore(); err != nil {
		return fmt.Errorf("failed to generate .gitignore: %w", err)
	}

	// Download dependencies and generate go.sum so the user can build
	// without network access to proxy.golang.org.
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
	binName := filepath.Base(g.outputDir)

	data := struct {
		ModuleName string
		BinaryName string
	}{
		ModuleName: moduleName,
		BinaryName: binName,
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

// ClientToolInfo holds the data needed to generate client examples for a single tool
type ClientToolInfo struct {
	Name         string
	Description  string
	Method       string
	ExampleArgs  string // JSON string ready for use in curl
}

// GenerateClientSh creates a client.sh script for quick manual testing
func (g *Generator) GenerateClientSh(config *converter.MCPConfig) error {
	clientTemplateContent, err := templatesFS.ReadFile("templates/client.sh.templ")
	if err != nil {
		return fmt.Errorf("failed to read client.sh template: %w", err)
	}

	tools := make([]ClientToolInfo, 0, len(config.Tools))
	limit := len(config.Tools)
	if limit > 20 {
		limit = 20
	}
	for _, tool := range config.Tools[:limit] {
		tools = append(tools, ClientToolInfo{
			Name:         capitalizeFirstLetter(tool.Name),
			Description:  tool.Description,
			Method:       tool.RequestTemplate.Method,
			ExampleArgs:  generateExampleArgs(tool),
		})
	}

	tmpl, err := template.New("client.sh").Funcs(template.FuncMap{
		"jsonExample": func(info ClientToolInfo) string { return info.ExampleArgs },
	}).Parse(string(clientTemplateContent))
	if err != nil {
		return fmt.Errorf("failed to parse client.sh template: %w", err)
	}

	data := struct {
		Tools []ClientToolInfo
	}{
		Tools: tools,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to render client.sh template: %w", err)
	}

	if err := writeFileContent(g.outputDir, "client.sh", func() ([]byte, error) {
		return buf.Bytes(), nil
	}); err != nil {
		return fmt.Errorf("failed to write client.sh file: %w", err)
	}

	// Make executable
	if err := os.Chmod(filepath.Join(g.outputDir, "client.sh"), 0755); err != nil {
		return fmt.Errorf("failed to chmod client.sh: %w", err)
	}

	return nil
}

// GenerateMakefile creates a Makefile for building and running the MCP server
func (g *Generator) GenerateMakefile() error {
	binName := filepath.Base(g.outputDir)
	makefile := fmt.Sprintf(".PHONY: build run clean test\n\nbuild:\n\t@go build -o %s .\n\nrun: build\n\t@./%s\n\nclean:\n\t@rm -f %s\n\ntest:\n\t@go test ./...\n", binName, binName, binName)

	if err := writeFileContent(g.outputDir, "Makefile", func() ([]byte, error) {
		return []byte(makefile), nil
	}); err != nil {
		return fmt.Errorf("failed to write Makefile: %w", err)
	}

	return nil
}

// GenerateReadme creates a README.md for the generated MCP server project
func (g *Generator) GenerateReadme() error {
	binName := filepath.Base(g.outputDir)
	mcpName := "myconfluence"

	readme := "# " + binName + "\n\n## Quick Start\n\n" +
		"```sh\nmake\n```\n\n" +
		"## Token Configuration\n\n" +
		"The server retrieves the upstream Bearer token using the following priority:\n\n" +
		"1. Authorization header from the client request (forwarded)\n" +
		"2. `MCP_UPSTREAM_TOKEN` environment variable\n" +
		"3. `.credentials` file (set `MCP_UPSTREAM_TOKEN_FILE=.credentials`)\n" +
		"4. macOS Keychain / Windows Credential Manager\n\n" +
		"To use the `.credentials` file:\n\n" +
		"```sh\necho -n \"your-token\" > .credentials\nexport MCP_UPSTREAM_TOKEN_FILE=.credentials\n```\n\n" +
		"## Agent Integration\n\n" +
		"### Local Mode (stdio)\n\n" +
		"Run the MCP server as a child process — recommended for local development.\n\n" +
		"### OpenCode\n\n" +
		"`~/.config/opencode/config.json`:\n\n" +
		"```json\n" +
		`{
  "mcp": {
    "` + mcpName + `": {
      "type": "local",
      "command": ["./` + binName + `"],
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://example.atlassian.net/wiki/rest/api",
        "MCP_UPSTREAM_TOKEN": "your-token"
      },
      "enabled": true
    }
  }
}
` + "```\n\n" +
		"### Claude Code\n\n" +
		"`~/.claude/settings.json`:\n\n" +
		"```json\n" +
		`{
  "mcpServers": {
    "` + mcpName + `": {
      "command": "./` + binName + `",
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://example.atlassian.net/wiki/rest/api",
        "MCP_UPSTREAM_TOKEN": "your-token"
      }
    }
  }
}
` + "```\n\n" +
		"### Claude Desktop\n\n" +
		"`~/.config/claude-desktop/claude_desktop_config.json`:\n\n" +
		"```json\n" +
		`{
  "mcpServers": {
    "` + mcpName + `": {
      "command": ["./` + binName + `"],
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://example.atlassian.net/wiki/rest/api",
        "MCP_UPSTREAM_TOKEN": "your-token"
      }
    }
  }
}
` + "```\n\n" +
		"### Codex CLI\n\n" +
		"`~/.codex/config.yaml`:\n\n" +
		"```yaml\n" +
		"mcp:\n" +
		"  servers:\n" +
		"    " + mcpName + ":\n" +
		"      command: ./" + binName + "\n" +
		"      args: [\"--transport\", \"stdio\"]\n" +
		"      env:\n" +
		"        MCP_UPSTREAM_ENDPOINT: https://example.atlassian.net/wiki/rest/api\n" +
		"        MCP_UPSTREAM_TOKEN: your-token\n" +
		"```\n\n" +
		"### Cursor\n\n" +
		"`~/.cursor/mcp.json`:\n\n" +
		"```json\n" +
		`{
  "mcpServers": {
    "` + mcpName + `": {
      "command": "./` + binName + `",
      "args": ["--transport", "stdio"],
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://example.atlassian.net/wiki/rest/api",
        "MCP_UPSTREAM_TOKEN": "your-token"
      }
    }
  }
}
` + "```\n\n" +
		"### Remote Mode (HTTP)\n\n" +
		"Run the server separately and connect agents via HTTP transport.\n\n" +
		"Start the server:\n\n" +
		"```sh\n" +
		"export MCP_UPSTREAM_ENDPOINT=https://example.atlassian.net/wiki/rest/api\n" +
		"export MCP_UPSTREAM_TOKEN=your-token\n" +
		"./" + binName + " --transport http --port 8080\n" +
		"```\n\n" +
		"### OpenCode (remote)\n\n" +
		"`~/.config/opencode/config.json`:\n\n" +
		"```json\n" +
		`{
  "mcp": {
    "` + mcpName + `": {
      "type": "remote",
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://example.atlassian.net/wiki/rest/api",
        "MCP_UPSTREAM_TOKEN": "your-token"
      }
    }
  }
}
` + "```\n\n" +
		"### Claude Code (remote)\n\n" +
		"`~/.claude/settings.json`:\n\n" +
		"```json\n" +
		`{
  "mcpServers": {
    "` + mcpName + `": {
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://example.atlassian.net/wiki/rest/api",
        "MCP_UPSTREAM_TOKEN": "your-token"
      }
    }
  }
}
` + "```\n\n" +
		"### Claude Desktop (remote)\n\n" +
		"`~/.config/claude-desktop/claude_desktop_config.json`:\n\n" +
		"```json\n" +
		`{
  "mcpServers": {
    "` + mcpName + `": {
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://example.atlassian.net/wiki/rest/api",
        "MCP_UPSTREAM_TOKEN": "your-token"
      }
    }
  }
}
` + "```\n\n" +
		"### Codex CLI (remote)\n\n" +
		"`~/.codex/config.yaml`:\n\n" +
		"```yaml\nmcp:\n  servers:\n    " + mcpName + ":\n      url: http://localhost:8080/mcp\n      env:\n        MCP_UPSTREAM_ENDPOINT: https://example.atlassian.net/wiki/rest/api\n        MCP_UPSTREAM_TOKEN: your-token\n```\n\n" +
		"### Cursor (remote)\n\n" +
		"`~/.cursor/mcp.json`:\n\n" +
		"```json\n" +
		`{
  "mcpServers": {
    "` + mcpName + `": {
      "url": "http://localhost:8080/mcp",
      "env": {
        "MCP_UPSTREAM_ENDPOINT": "https://example.atlassian.net/wiki/rest/api",
        "MCP_UPSTREAM_TOKEN": "your-token"
      }
    }
  }
}
` + "```\n"

	if err := writeFileContent(g.outputDir, "README.md", func() ([]byte, error) {
		return []byte(readme), nil
	}); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	return nil
}

// GenerateDotCredentials creates a .credentials file for storing the upstream token.
// The generated MCP server can read the token from this file via MCP_UPSTREAM_TOKEN_FILE.
func (g *Generator) GenerateDotCredentials() error {
	// Only create if it doesn't already exist (preserve user's token)
	path := filepath.Join(g.outputDir, ".credentials")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.WriteFile(path, []byte(""), 0600); err != nil {
		return fmt.Errorf("failed to write .credentials: %w", err)
	}
	return nil
}

// GenerateDotGitignore creates a .gitignore for the generated MCP server project.
func (g *Generator) GenerateDotGitignore() error {
	content := ".credentials\n"
	if err := writeFileContent(g.outputDir, ".gitignore", func() ([]byte, error) {
		return []byte(content), nil
	}); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}
	return nil
}

type argEntry struct {
	key   string
	value string
}

// generateExampleArgs builds a JSON args string from a tool's schema.
// It picks example values or defaults from the schema, falling back to sensible type-based defaults.
func generateExampleArgs(tool converter.Tool) string {
	var topArgs []argEntry
	var bodyArgs []argEntry

	for _, arg := range tool.Args {
		if arg.Source == "body" && len(arg.ContentTypes) > 0 {
			// Use the JSON content type schema (prefer application/json)
			var jsonSchema *converter.Schema
			if s, ok := arg.ContentTypes["application/json"]; ok {
				jsonSchema = s
			} else {
				for _, s := range arg.ContentTypes {
					jsonSchema = s
					break
				}
			}
			if jsonSchema != nil && jsonSchema.Object != nil {
				var bodyEntries []argEntry
				for propName, propSchema := range jsonSchema.Object.Properties {
					if propSchema.ReadOnly {
						continue
					}
					val := argValueFromSchema(propName, propSchema)
					bodyEntries = append(bodyEntries, argEntry{key: propName, value: val})
				}
				if len(bodyEntries) > 0 {
					bodyArgs = append(bodyArgs, argEntry{key: arg.Name, value: buildArgsObject(bodyEntries)})
				} else {
					bodyArgs = append(bodyArgs, argEntry{key: arg.Name, value: "{}"})
				}
			} else {
				bodyArgs = append(bodyArgs, argEntry{key: arg.Name, value: argValue(arg)})
			}
		} else {
			val := argValue(arg)
			topArgs = append(topArgs, argEntry{key: arg.Name, value: val})
		}
	}

		// Always include body args in the example (input schema always has "body" property)
		if len(bodyArgs) > 0 {
			// bodyArgs entries have key="body" and value=body-object-JSON, use the value directly
			topArgs = append(topArgs, argEntry{key: "body", value: bodyArgs[0].value})
		}

	return buildArgsObject(topArgs)
}

func buildArgsObject(entries []argEntry) string {
	if len(entries) == 0 {
		return "{}"
	}
	var b strings.Builder
	b.WriteString("{")
	for i, e := range entries {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("\"")
		b.WriteString(e.key)
		b.WriteString("\": ")
		b.WriteString(e.value)
	}
	b.WriteString("}")
	return b.String()
}

func argValue(arg converter.Arg) string {
	// 1. Use schema example if available
	if arg.Schema != nil && arg.Schema.Example != nil {
		return jsonEncode(arg.Schema.Example)
	}

	// 2. Use default if available
	if arg.Schema != nil && arg.Schema.Default != nil {
		return jsonEncode(arg.Schema.Default)
	}

	// 3. Use enum first value if available
	if arg.Schema != nil && len(arg.Schema.Enum) > 0 {
		return jsonEncode(arg.Schema.Enum[0])
	}

	// 4. Fall back to type-based defaults
	if arg.Schema != nil && len(arg.Schema.Types) > 0 {
		t := arg.Schema.Types[0]
		switch t {
		case "string":
			if arg.Schema.Format == "uuid" {
				return `"550e8400-e29b-41d4-a716-446655440000"`
			}
			if arg.Schema.Format == "date" || arg.Schema.Format == "date-time" {
				return `"2025-01-01"`
			}
			if arg.Schema.Format == "email" {
				return `"user@example.com"`
			}
			// Use description or name for context
			if arg.Schema.Description != "" {
				return fmt.Sprintf(`"%s_value"`, arg.Name)
			}
			return fmt.Sprintf(`"%s_value"`, arg.Name)
		case "integer", "number":
			return "0"
		case "boolean":
			return "false"
		case "array":
			return "[]"
		case "object":
			return "{}"
		}
	}

	// 5. Last resort
	return `"value"`
}

func jsonEncode(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// argValueFromSchema generates an example value directly from a Schema (for nested properties).
func argValueFromSchema(name string, s *converter.Schema) string {
	if s.Example != nil {
		return jsonEncode(s.Example)
	}
	if s.Default != nil {
		return jsonEncode(s.Default)
	}
	if len(s.Enum) > 0 {
		return jsonEncode(s.Enum[0])
	}
	if len(s.Types) > 0 {
		t := s.Types[0]
		switch t {
		case "string":
			if s.Format == "uuid" {
				return `"550e8400-e29b-41d4-a716-446655440000"`
			}
			if s.Format == "date" || s.Format == "date-time" {
				return `"2025-01-01"`
			}
			if s.Format == "email" {
				return `"user@example.com"`
			}
			return fmt.Sprintf(`"%s_value"`, name)
		case "integer", "number":
			return "0"
		case "boolean":
			return "false"
		case "array":
			return "[]"
		case "object":
			return "{}"
		}
	}
	return `"value"`
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

// RunGoBuild compiles the MCP server binary in the output directory
func (g *Generator) RunGoBuild() error {
	binName := filepath.Base(g.outputDir)
	binPath := filepath.Join(g.outputDir, binName)
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = g.outputDir
	cmd.Env = append(os.Environ(), "GOPROXY=https://proxy.golang.org,direct", "GONOSUMCHECK=*", "GOSUMDB=off")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	return nil
}
