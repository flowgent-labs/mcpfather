package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// GenerateHelpers creates a client.go file with utility functions for MCP tools
func (g *Generator) GenerateHelpers() error {
	if err := g.generateClientGo(); err != nil {
		return err
	}
	return g.generateRequestLog()
}

// generateClientGo creates the client.go file (ForwardRequest, params helpers)
func (g *Generator) generateClientGo() error {
	helpersTemplate, err := templatesFS.ReadFile("templates/helpers.templ")
	if err != nil {
		return fmt.Errorf("failed to read helpers template file: %w", err)
	}

	tmpl, err := template.New("helpers").Parse(string(helpersTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse helpers template: %w", err)
	}

	data := struct{}{}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return fmt.Errorf("failed to execute helpers template: %w", err)
	}

	formattedCode, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated helpers code: %w", err)
	}

	err = writeFileContent(g.outputDir+"/internal/helpers", "client.go", func() ([]byte, error) {
		return formattedCode, nil
	})
	if err != nil {
		return fmt.Errorf("failed to write helpers.go file: %w", err)
	}

	// Remove old params.go if it exists from a previous generation
	oldFile := filepath.Join(g.outputDir, "internal", "helpers", "params.go")
	os.Remove(oldFile)

	return nil
}

// generateRequestLog creates the request_log.go file with kubectl-style verbosity logging
func (g *Generator) generateRequestLog() error {
	reqLogTemplate, err := templatesFS.ReadFile("templates/request_log.templ")
	if err != nil {
		return fmt.Errorf("failed to read request_log template file: %w", err)
	}

	tmpl, err := template.New("request_log").Parse(string(reqLogTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse request_log template: %w", err)
	}

	data := struct{}{}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return fmt.Errorf("failed to execute request_log template: %w", err)
	}

	formattedCode, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated request_log code: %w", err)
	}

	err = writeFileContent(g.outputDir+"/internal/helpers", "request_log.go", func() ([]byte, error) {
		return formattedCode, nil
	})
	if err != nil {
		return fmt.Errorf("failed to write request_log.go file: %w", err)
	}

	return nil
}

// GenerateCredentials copies credential manager files to the helpers package.
// These files use Go build tags to support macOS Keychain, Windows Credential
// Manager, and provide stubs for other platforms.
func (g *Generator) GenerateCredentials() error {
	credFiles := []string{
		"token_keychain.go",
		"token_wincred.go",
		"token_other.go",
	}
	for _, f := range credFiles {
		content, err := templatesFS.ReadFile("templates/_credentials/" + f)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", f, err)
		}
		if err := writeFileContent(g.outputDir+"/internal/helpers", f, func() ([]byte, error) {
			return content, nil
		}); err != nil {
			return fmt.Errorf("failed to write %s: %w", f, err)
		}
	}
	return nil
}
