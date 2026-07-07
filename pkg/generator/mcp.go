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

	"github.com/flowgent-labs/mcpfather/pkg/converter"
)

//go:embed templates/*.templ
//go:embed mcpvirtual/*/*.go
//go:embed skills/virtual-tool-creator
//go:embed deploy/helm/*
//go:embed deploy/helm/templates/*
//go:embed deploy/docker/*
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
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] starting MCP generation\n")
	}

	config, err := g.converter.Convert()
	if err != nil {
		return fmt.Errorf("failed at converting OpenAPI schema into MCP code %w", err)
	}
	g.tools = config.Tools

	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] conversion complete: %d tools generated\n", len(config.Tools))
	}

	if err := GenerateGoMod(g.outputDir); err != nil {
		return fmt.Errorf("failed to generate go.mod: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated go.mod\n")
	}

	if err := g.GenerateMainGo(); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated main.go\n")
	}

	if err := g.GenerateServerFile(config); err != nil {
		return fmt.Errorf("failed to generate server file: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated pkg/mcpserver/server.go\n")
	}

	if err := g.GenerateToolFiles(config); err != nil {
		return fmt.Errorf("failed to generate tool files: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated %d tool files in pkg/mcptools/\n", len(config.Tools))
	}

	if err := g.GenerateToolRegistry(config); err != nil {
		return fmt.Errorf("failed to generate tool registry: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated pkg/mcptools/registry.go\n")
	}

	if err := g.GenerateCLI(); err != nil {
		return fmt.Errorf("failed to generate CLI package: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated pkg/mcpcli/cli.go\n")
	}

	if err := g.GenerateHelpers(); err != nil {
		return fmt.Errorf("failed to generate helpers: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated pkg/helpers/\n")
	}

	if err := g.GenerateMetrics(); err != nil {
		return fmt.Errorf("failed to generate metrics: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated pkg/helpers/metrics.go\n")
	}

	if err := g.GenerateTrace(); err != nil {
		return fmt.Errorf("failed to generate trace: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated pkg/helpers/trace.go\n")
	}

	if err := g.GenerateVirtual(); err != nil {
		return fmt.Errorf("failed to generate virtual engine: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated pkg/mcpvirtual/\n")
	}

	if err := g.GenerateVirtualToolCreatorSkill(); err != nil {
		return fmt.Errorf("failed to generate virtual-tool-creator skill: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated .agents/skills/virtual-tool-creator/\n")
	}

	if err := g.GenerateClientSh(config); err != nil {
		return fmt.Errorf("failed to generate client script: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated mcpclient.sh\n")
	}

	if err := g.GenerateMakefile(); err != nil {
		return fmt.Errorf("failed to generate Makefile: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated Makefile\n")
	}

	if err := g.GenerateReadme(); err != nil {
		return fmt.Errorf("failed to generate README.md: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated README.md\n")
	}

	if err := g.GenerateDotCredentials(); err != nil {
		return fmt.Errorf("failed to generate .credentials: %w", err)
	}

	if err := g.GenerateDotGitignore(); err != nil {
		return fmt.Errorf("failed to generate .gitignore: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated .gitignore\n")
	}

	if err := g.GenerateDeploy(); err != nil {
		return fmt.Errorf("failed to generate deploy/: %w", err)
	}
	if g.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] generated deploy/\n")
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
	Name        string
	Description string
	Method      string
	ExampleArgs string
	UploadCT    string // non-empty if this is an upload tool
}

// GenerateClientSh creates a mcpclient.sh script for quick manual testing
func (g *Generator) GenerateClientSh(config *converter.MCPConfig) error {
	clientTemplateContent, err := templatesFS.ReadFile("templates/mcpclient.sh.templ")
	if err != nil {
		return fmt.Errorf("failed to read mcpclient.sh template: %w", err)
	}

	tools := make([]ClientToolInfo, 0, len(config.Tools))
	limit := len(config.Tools)
	if limit > 20 {
		limit = 20
	}
	for _, tool := range config.Tools[:limit] {
		info := ClientToolInfo{
			Name:        capitalizeFirstLetter(tool.Name),
			Description: tool.Description,
			Method:      tool.RequestTemplate.Method,
			ExampleArgs: generateExampleArgs(tool),
			UploadCT:    tool.UploadContentType,
		}
		tools = append(tools, info)
	}

	tmpl, err := template.New("mcpclient.sh").Funcs(template.FuncMap{
		"jsonExample": func(info ClientToolInfo) string { return info.ExampleArgs },
	}).Parse(string(clientTemplateContent))
	if err != nil {
		return fmt.Errorf("failed to parse mcpclient.sh template: %w", err)
	}

	data := struct {
		Tools []ClientToolInfo
	}{
		Tools: tools,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to render mcpclient.sh template: %w", err)
	}

	if err := writeFileContent(g.outputDir, "mcpclient.sh", func() ([]byte, error) {
		return buf.Bytes(), nil
	}); err != nil {
		return fmt.Errorf("failed to write mcpclient.sh file: %w", err)
	}

	// Make executable
	if err := os.Chmod(filepath.Join(g.outputDir, "mcpclient.sh"), 0755); err != nil {
		return fmt.Errorf("failed to chmod mcpclient.sh: %w", err)
	}

	return nil
}

// GenerateMakefile creates a Makefile for building and running the MCP server
func (g *Generator) GenerateMakefile() error {
	binName := filepath.Base(g.outputDir)
	makefile := fmt.Sprintf(`# Generated by https://github.com/flowgent-labs/mcpfather

.PHONY: build build-all run clean test

BINARY_NAME := %s
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS := -v -trimpath

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

BIN := bin/$(BINARY_NAME)-$(GOOS)-$(GOARCH)-$(VERSION)$(if $(filter windows,$(GOOS)),.exe,)

build: go.sum
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -o $(BIN) .
	@ln -sf $(notdir $(BIN)) bin/$(BINARY_NAME)

build-all: go.sum
	GOOS=linux   GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-linux-amd64-$(VERSION)   .
	GOOS=linux   GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-linux-arm64-$(VERSION)   .
	GOOS=darwin  GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-amd64-$(VERSION)  .
	GOOS=darwin  GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-arm64-$(VERSION)  .
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-windows-amd64-$(VERSION).exe .
	GOOS=windows GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-windows-arm64-$(VERSION).exe .

go.sum: go.mod
	go mod tidy

run: build
	@bin/$(BINARY_NAME)

clean:
	@rm -f bin/$(BINARY_NAME)*

test:
	@go test ./...

# ---- Optional: OpenTelemetry distributed tracing ----
# Build with -tags otel to include OTLP gRPC tracing support
# (Prometheus metrics are always compiled in by default).

build-with-otel: go.sum
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -tags otel $(BUILD_FLAGS) -o $(BIN) .
	@ln -sf $(notdir $(BIN)) bin/$(BINARY_NAME)

# ---- Container & Kubernetes ----

IMAGE_REPO ?= docker.io/library/$(BINARY_NAME)
IMAGE_TAG  ?= $(VERSION)
MCP_UPSTREAM_ENDPOINT  ?= $(MCP_UPSTREAM_ENDPOINT)
MCP_UPSTREAM_TOKEN  ?= $(MCP_UPSTREAM_TOKEN)
MCP_OIDC_ISSUER_URL  ?= $(MCP_OIDC_ISSUER_URL)
MCP_OIDC_CLIENT_ID  ?= $(MCP_OIDC_CLIENT_ID)
MCP_OIDC_CLIENT_SECRET  ?= $(MCP_OIDC_CLIENT_SECRET)

build-image: build
	docker build -t $(IMAGE_REPO):$(IMAGE_TAG) -f deploy/docker/Dockerfile .

build-image-with-otel: build-with-otel
	docker build --build-arg BUILD_TAGS=otel -t $(IMAGE_REPO):$(IMAGE_TAG)-otel -f deploy/docker/Dockerfile .

deploy: build-image
	helm upgrade -i $(BINARY_NAME) deploy/helm \
		--set image.repository=$(IMAGE_REPO) \
		--set image.tag=$(IMAGE_TAG) \
		--set config.upstream.endpoint=$(MCP_UPSTREAM_ENDPOINT) \
		--set secret.static.create=true \
		--set secret.static.bearerToken=$(MCP_UPSTREAM_TOKEN) \
		--set secret.oidc.enabled=true \
		--set secret.oidc.issuerUrl=$(MCP_OIDC_ISSUER_URL) \
		--set secret.oidc.clientId=$(MCP_OIDC_CLIENT_ID) \
		--set secret.oidc.clientSecret=$(MCP_OIDC_CLIENT_SECRET)
`, binName)

	if err := writeFileContent(g.outputDir, "Makefile", func() ([]byte, error) {
		return []byte(makefile), nil
	}); err != nil {
		return fmt.Errorf("failed to write Makefile: %w", err)
	}

	return nil
}

// GenerateToolRegistry creates a registry.go file in the mcptools package that
// maps tool names to their Tool and Handler, allowing dynamic tool discovery
// by both the MCP server and the CLI runner.
func (g *Generator) GenerateToolRegistry(config *converter.MCPConfig) error {
	var buf bytes.Buffer
	buf.WriteString("// Code generated by https://github.com/flowgent-labs/mcpfather\n\n")
	buf.WriteString("package mcptools\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"context\"\n\n")
	buf.WriteString("\t\"github.com/mark3labs/mcp-go/mcp\"\n")
	buf.WriteString(")\n\n")
	buf.WriteString("// ToolEntry pairs an MCP tool definition with its handler function.\n")
	buf.WriteString("type ToolEntry struct {\n")
	buf.WriteString("\tTool    mcp.Tool\n")
	buf.WriteString("\tHandler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)\n")
	buf.WriteString("}\n\n")
	buf.WriteString("// Registry maps tool names to their ToolEntry for dynamic tool discovery.\n")
	buf.WriteString("var Registry = map[string]ToolEntry{\n")

	for _, tool := range config.Tools {
		capitalizedName := capitalizeFirstLetter(tool.Name)
		fmt.Fprintf(&buf, "\t%q: {Tool: New%sMCPTool(), Handler: %sHandler},\n",
			capitalizedName, capitalizedName, capitalizedName)
	}

	buf.WriteString("}\n")

	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated registry.go: %w", err)
	}

	if err := writeFileContent(g.outputDir+"/pkg/mcptools", "registry.go", func() ([]byte, error) {
		return formattedCode, nil
	}); err != nil {
		return fmt.Errorf("failed to write registry.go: %w", err)
	}

	return nil
}

// GenerateCLI creates the mcpcli package that provides a CLI interface
// for dynamically invoking MCP tools from the command line.
func (g *Generator) GenerateCLI() error {
	cliTemplateContent, err := templatesFS.ReadFile("templates/cli.templ")
	if err != nil {
		return fmt.Errorf("failed to read CLI template: %w", err)
	}

	tmpl, err := template.New("cli.templ").Parse(string(cliTemplateContent))
	if err != nil {
		return fmt.Errorf("failed to parse CLI template: %w", err)
	}

	importPath, err := BuildImportPath(g.outputDir)
	if err != nil {
		return fmt.Errorf("failed to build import path: %w", err)
	}

	helpersImportPath := BuildModuleName(g.outputDir) + "/pkg/helpers"
	virtualImportPath := BuildModuleName(g.outputDir) + "/pkg/mcpvirtual"

	data := struct {
		MCPToolsImportPath string
		HelpersImportPath  string
		VirtualImportPath  string
		BinaryName         string
	}{
		MCPToolsImportPath: importPath,
		HelpersImportPath:  helpersImportPath,
		VirtualImportPath:  virtualImportPath,
		BinaryName:         filepath.Base(g.outputDir),
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to render CLI template: %w", err)
	}

	formattedCode, err := format.Source(body.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated cli.go: %w", err)
	}

	if err := writeFileContent(g.outputDir+"/pkg/mcpcli", "cli.go", func() ([]byte, error) {
		return formattedCode, nil
	}); err != nil {
		return fmt.Errorf("failed to write cli.go: %w", err)
	}

	return nil
}

// readmeToolEntry holds a single tool entry for the README tool table.
type readmeToolEntry struct {
	Name        string
	Description string
}

// readmeTemplateData is the data passed to readme.templ.
type readmeTemplateData struct {
	BinaryName     string
	ToolCount      int
	Tools          []readmeToolEntry
	RemainingCount int
}

// GenerateReadme creates a README.md for the generated MCP server project.
func (g *Generator) GenerateReadme() error {
	binName := filepath.Base(g.outputDir)

	tools := make([]readmeToolEntry, 0, len(g.tools))
	limit := 15
	for i, t := range g.tools {
		if i >= limit {
			break
		}
		desc := strings.ReplaceAll(t.Description, "\n", " ")
		desc = strings.ReplaceAll(desc, "\r", "")
		desc = strings.ReplaceAll(desc, "|", "\\|")
		tools = append(tools, readmeToolEntry{
			Name:        t.Name,
			Description: desc,
		})
	}

	data := readmeTemplateData{
		BinaryName:     binName,
		ToolCount:      len(g.tools),
		Tools:          tools,
		RemainingCount: len(g.tools) - limit,
	}

	tmplContent, err := templatesFS.ReadFile("templates/readme.templ")
	if err != nil {
		return fmt.Errorf("failed to read readme template: %w", err)
	}

	tmpl, err := template.New("readme.templ").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse readme template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute readme template: %w", err)
	}

	if err := os.WriteFile(filepath.Join(g.outputDir, "README.md"), buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	return nil
}

// GenerateDotCredentials creates a .credentials file for storing the upstream token.
// The generated MCP server can read the token from this file via MCP__AUTH__STATIC__BEARER_TOKEN_FILE.
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
	content := "# Generated by https://github.com/flowgent-labs/mcpfather\n\n.credentials\nbin/\n*.exe\n*.dll\n*.so\n*.dylib\n*.test\n*.out\n"
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
