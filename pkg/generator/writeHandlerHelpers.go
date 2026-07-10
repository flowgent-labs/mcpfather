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
	if err := g.generateResourceServerGo(); err != nil {
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

	data := struct{ ModuleName string }{ModuleName: BuildModuleName(g.outputDir)}
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
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

// generateResourceServerGo creates the resource_server.go file (inbound JWT
// bearer token validation — the MCP server's Resource Server role).
func (g *Generator) generateResourceServerGo() error {
	t, err := templatesFS.ReadFile("templates/resource_server.templ")
	if err != nil {
		return fmt.Errorf("failed to read resource_server template file: %w", err)
	}

	tmpl, err := template.New("resource_server").Parse(string(t))
	if err != nil {
		return fmt.Errorf("failed to parse resource_server template: %w", err)
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, nil); err != nil {
		return fmt.Errorf("failed to execute resource_server template: %w", err)
	}

	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format generated resource_server code: %w\n%s", err, buffer.String())
	}

	return writeFileContent(g.outputDir+"/pkg/helpers", "resource_server.go", func() ([]byte, error) {
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

// GenerateTrace creates trace_grpc.go, trace_http.go, and trace_noop.go with
// OpenTelemetry tracing (OTLP export).
//
//	trace_grpc.go  (build tag: otel_grpc) — OTLP gRPC exporter
//	trace_http.go  (build tag: otel_http) — OTLP HTTP/protobuf exporter
//	trace_noop.go  (default, no build tag) — stubs compiled by default
//
// Use -tags otel_grpc or -tags otel_http to enable distributed tracing.
func (g *Generator) GenerateTrace() error {
	// trace_grpc.go
	grpcTemplate, err := templatesFS.ReadFile("templates/trace_grpc.templ")
	if err != nil {
		return fmt.Errorf("failed to read trace_grpc template: %w", err)
	}

	grpcTmpl, err := template.New("trace_grpc").Parse(string(grpcTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse trace_grpc template: %w", err)
	}

	var grpcBuf bytes.Buffer
	if err := grpcTmpl.Execute(&grpcBuf, struct{}{}); err != nil {
		return fmt.Errorf("failed to execute trace_grpc template: %w", err)
	}

	grpcFormatted, err := format.Source(grpcBuf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format trace_grpc: %w\n%s", err, grpcBuf.String())
	}

	if err := writeFileContent(g.outputDir+"/pkg/helpers", "trace_grpc.go", func() ([]byte, error) {
		return grpcFormatted, nil
	}); err != nil {
		return fmt.Errorf("failed to write trace_grpc.go: %w", err)
	}

	// trace_http.go
	httpTemplate, err := templatesFS.ReadFile("templates/trace_http.templ")
	if err != nil {
		return fmt.Errorf("failed to read trace_http template: %w", err)
	}

	httpTmpl, err := template.New("trace_http").Parse(string(httpTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse trace_http template: %w", err)
	}

	var httpBuf bytes.Buffer
	if err := httpTmpl.Execute(&httpBuf, struct{}{}); err != nil {
		return fmt.Errorf("failed to execute trace_http template: %w", err)
	}

	httpFormatted, err := format.Source(httpBuf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format trace_http: %w\n%s", err, httpBuf.String())
	}

	if err := writeFileContent(g.outputDir+"/pkg/helpers", "trace_http.go", func() ([]byte, error) {
		return httpFormatted, nil
	}); err != nil {
		return fmt.Errorf("failed to write trace_http.go: %w", err)
	}

	// trace_noop.go (default, no build tag)
	noopTemplate, err := templatesFS.ReadFile("templates/trace_noop.templ")
	if err != nil {
		return fmt.Errorf("failed to read trace_noop template: %w", err)
	}

	noopTmpl, err := template.New("trace_noop").Parse(string(noopTemplate))
	if err != nil {
		return fmt.Errorf("failed to parse trace_noop template: %w", err)
	}

	var noopBuf bytes.Buffer
	if err := noopTmpl.Execute(&noopBuf, struct{}{}); err != nil {
		return fmt.Errorf("failed to execute trace_noop template: %w", err)
	}

	noopFormatted, err := format.Source(noopBuf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format trace_noop: %w\n%s", err, noopBuf.String())
	}

	if err := writeFileContent(g.outputDir+"/pkg/helpers", "trace_noop.go", func() ([]byte, error) {
		return noopFormatted, nil
	}); err != nil {
		return fmt.Errorf("failed to write trace_noop.go: %w", err)
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

