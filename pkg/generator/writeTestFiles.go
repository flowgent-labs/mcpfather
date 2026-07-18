package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// GenerateCommonTestFiles writes common _test.go files for all shared packages.
func (g *Generator) GenerateCommonTestFiles() error {
	moduleName := BuildModuleName(g.outputDir)
	binName := filepath.Base(g.outputDir)

	data := struct {
		ModuleName string
		BinaryName string
	}{ModuleName: moduleName, BinaryName: binName}

	templatesToGenerate := []struct {
		tmplName string
		outDir   string
		outFile  string
	}{
		{"helpers_client_test.templ", "pkg/helpers", "client_test.go"},
		{"helpers_config_test.templ", "pkg/helpers", "config_test.go"},
		{"helpers_resource_server_test.templ", "pkg/helpers", "resource_server_test.go"},
		{"mcpconfig_types_test.templ", "pkg/mcpconfig", "types_test.go"},
		{"registry_test.templ", "pkg/mcptools", "registry_test.go"},
		{"server_test.templ", "pkg/mcpserver", "server_test.go"},
	}

	for _, tt := range templatesToGenerate {
		content, err := templatesFS.ReadFile("templates/" + tt.tmplName)
		if err != nil {
			return fmt.Errorf("read template %s: %w", tt.tmplName, err)
		}

		tmpl, err := template.New(tt.tmplName).Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", tt.tmplName, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("render template %s: %w", tt.tmplName, err)
		}

		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			dumpPath := filepath.Join(g.outputDir, tt.outDir, tt.outFile+".bad")
			_ = os.WriteFile(dumpPath, buf.Bytes(), 0644)
			return fmt.Errorf("format %s: %w (raw written to %s)", tt.outFile, err, dumpPath)
		}

		outDir := filepath.Join(g.outputDir, tt.outDir)
		if err := writeFileContent(outDir, tt.outFile, func() ([]byte, error) {
			return formatted, nil
		}); err != nil {
			return fmt.Errorf("write %s: %w", tt.outFile, err)
		}
	}

	return nil
}
