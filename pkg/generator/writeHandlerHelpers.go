package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

// GenerateHelpers creates the helpers package with all utility files.
func (g *Generator) GenerateHelpers() error {
	if err := g.generateConfigGo(); err != nil {
		return err
	}
	if err := g.generateAuthGo(); err != nil {
		return err
	}
	if err := g.generateClientGo(); err != nil {
		return err
	}
	return g.generateRequestLog()
}

// generateConfigGo creates the config.go file (Config structs, viper loading, env override).
func (g *Generator) generateConfigGo() error {
	t, err := templatesFS.ReadFile("templates/config.templ")
	if err != nil {
		return fmt.Errorf("failed to read config template file: %w", err)
	}

	tmpl, err := template.New("config").Parse(string(t))
	if err != nil {
		return fmt.Errorf("failed to parse config template: %w", err)
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, nil); err != nil {
		return fmt.Errorf("failed to execute config template: %w", err)
	}

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated config code: %w\n%s", err, buffer.String())
	}

	return writeFileContent(g.outputDir+"/pkg/helpers", "config.go", func() ([]byte, error) {
		return formatted, nil
	})
}

// generateAuthGo creates the auth.go file (OIDC token manager, static auth).
func (g *Generator) generateAuthGo() error {
	t, err := templatesFS.ReadFile("templates/auth.templ")
	if err != nil {
		return fmt.Errorf("failed to read auth template file: %w", err)
	}

	tmpl, err := template.New("auth").Parse(string(t))
	if err != nil {
		return fmt.Errorf("failed to parse auth template: %w", err)
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, nil); err != nil {
		return fmt.Errorf("failed to execute auth template: %w", err)
	}

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated auth code: %w\n%s", err, buffer.String())
	}

	return writeFileContent(g.outputDir+"/pkg/helpers", "auth.go", func() ([]byte, error) {
		return formatted, nil
	})
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

	err = writeFileContent(g.outputDir+"/pkg/helpers", "client.go", func() ([]byte, error) {
		return formattedCode, nil
	})
	if err != nil {
		return fmt.Errorf("failed to write helpers.go file: %w", err)
	}

	// Remove old params.go if it exists from a previous generation
	oldFile := filepath.Join(g.outputDir, "pkg", "helpers", "params.go")
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

	err = writeFileContent(g.outputDir+"/pkg/helpers", "request_log.go", func() ([]byte, error) {
		return formattedCode, nil
	})
	if err != nil {
		return fmt.Errorf("failed to write request_log.go file: %w", err)
	}

	return nil
}

// GenerateTrace creates trace.go and trace_noop.go with OpenTelemetry tracing (OTLP export).
// trace.go (build tag: otel) carries the full OTel SDK + gRPC dependency.
// trace_noop.go (default, no build tag) provides stubs compiled by default.
// Use -tags otel to enable distributed tracing.
func (g *Generator) GenerateTrace() error {
	traceTemplate, err := templatesFS.ReadFile("templates/trace.templ")
	if err != nil {
		return fmt.Errorf("failed to read trace template file: %w", err)
	}

	tmpl, err := template.New("trace").Parse(string(traceTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse trace template: %w", err)
	}

	data := struct{}{}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return fmt.Errorf("failed to execute trace template: %w", err)
	}

	formattedCode, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated trace code: %w", err)
	}

	err = writeFileContent(g.outputDir+"/pkg/helpers", "trace.go", func() ([]byte, error) {
		return formattedCode, nil
	})
	if err != nil {
		return fmt.Errorf("failed to write trace.go file: %w", err)
	}

	// No-op stub for builds with -tags no_otel
	traceNoopTemplate, err := templatesFS.ReadFile("templates/trace_noop.templ")
	if err != nil {
		return fmt.Errorf("failed to read trace_noop template file: %w", err)
	}

	tmplNoop, err := template.New("trace_noop").Parse(string(traceNoopTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse trace_noop template: %w", err)
	}

	var bufferNoop bytes.Buffer
	if err := tmplNoop.Execute(&bufferNoop, data); err != nil {
		return fmt.Errorf("failed to execute trace_noop template: %w", err)
	}

	formattedNoop, err := format.Source(bufferNoop.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated trace_noop code: %w", err)
	}

	err = writeFileContent(g.outputDir+"/pkg/helpers", "trace_noop.go", func() ([]byte, error) {
		return formattedNoop, nil
	})
	if err != nil {
		return fmt.Errorf("failed to write trace_noop.go file: %w", err)
	}

	return nil
}

// GenerateMetrics creates metrics.go with OpenTelemetry instrumentation for tool calls.
// Prometheus metrics are always compiled in — they are a core built-in capability.
// Use -tags otel to additionally enable OpenTelemetry distributed tracing.
func (g *Generator) GenerateMetrics() error {
	metricsTemplate, err := templatesFS.ReadFile("templates/metrics.templ")
	if err != nil {
		return fmt.Errorf("failed to read metrics template file: %w", err)
	}

	tmpl, err := template.New("metrics").Parse(string(metricsTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse metrics template: %w", err)
	}

	data := struct{}{}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return fmt.Errorf("failed to execute metrics template: %w", err)
	}

	formattedCode, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated metrics code: %w", err)
	}

	err = writeFileContent(g.outputDir+"/pkg/helpers", "metrics.go", func() ([]byte, error) {
		return formattedCode, nil
	})
	if err != nil {
		return fmt.Errorf("failed to write metrics.go file: %w", err)
	}

	return nil
}

// (Legacy credential helpers removed — keychain/wincred stubs are in config.templ)

