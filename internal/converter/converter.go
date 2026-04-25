package converter

import (
	"fmt"
	"sort"
	"strings"
)

// Converter represents an OpenAPI to MCP converter
type Converter struct {
	parser  *Parser
	options ConvertOptions
}

type ConverterInterface interface {
	Convert() (*MCPConfig, error)
}

// NewConverter creates a new OpenAPI to MCP converter
func NewConverter(parser *Parser, includePaths []string, excludePaths []string) (*Converter, error) {
	includeSet := make(map[string]struct{})
	excludeSet := make(map[string]struct{})

	for _, p := range includePaths {
		p = cleanFilterPath(p)
		if p != "" {
			includeSet[p] = struct{}{}
		}
	}
	for _, p := range excludePaths {
		p = cleanFilterPath(p)
		if p != "" {
			excludeSet[p] = struct{}{}
		}
	}

	// Check for conflicts
	for p := range includeSet {
		if _, ok := excludeSet[p]; ok {
			return nil, fmt.Errorf("path '%s' is specified in both --includes and --excludes", p)
		}
	}

	return &Converter{
		parser: parser,
		options: ConvertOptions{
			ServerConfig: make(map[string]interface{}),
			IncludePaths: includeSet,
			ExcludePaths: excludeSet,
		},
	}, nil
}


// Convert converts an OpenAPI document to an MCP configuration
func (c *Converter) Convert() (*MCPConfig, error) {
	if c.parser.GetDocument() == nil {
		return nil, fmt.Errorf("no OpenAPI document loaded")
	}

	// Create the MCP configuration
	config := &MCPConfig{
		Server: ServerConfig{
			Config: c.options.ServerConfig,
		},
		Tools: []Tool{},
	}

	// Process each path and operation
	for path, pathItem := range c.parser.GetPaths() {
		operations := getOperations(pathItem)
		for method, operation := range operations {
			if !c.shouldIncludePath(path, method) {
				continue
			}

			tool, err := c.convertOperation(path, method, operation)
			if err != nil {
				return nil, fmt.Errorf("failed to convert operation %s %s: %w", method, path, err)
			}
			config.Tools = append(config.Tools, *tool)
		}
	}

	// Sort tools by name for consistent output
	sort.Slice(config.Tools, func(i, j int) bool {
		return config.Tools[i].Name < config.Tools[j].Name
	})

	return config, nil
}

// cleanFilterPath strips surrounding quotes and whitespace from a filter path
// entry to handle values copied from PowerShell, bash, or YAML.
func cleanFilterPath(p string) string {
	p = strings.TrimSpace(p)
	// Remove surrounding quotes (PowerShell/bash)
	if len(p) >= 2 && ((p[0] == '"' && p[len(p)-1] == '"') || (p[0] == '\'' && p[len(p)-1] == '\'')) {
		p = p[1 : len(p)-1]
	}
	return strings.TrimSpace(p)
}

// normalizePath strips trailing slashes/colons, removes BOM and invisible
// characters, and lowercases for consistent matching.
func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	// Remove BOM and other invisible characters that may come from
	// copying on Windows (e.g., from YAML editors or terminals)
	p = strings.Trim(p, "\xef\xbb\xbf\uFEFF")
	p = strings.TrimRight(p, "/:")
	if p == "" {
		p = "/"
	}
	return strings.ToLower(p)
}

// normalizePathParam replaces all path variable segments with a generic
// placeholder so that /api/v2/scans/{scan_id} matches /api/v2/scans/{id}.
func normalizePathParam(p string) string {
	p = normalizePath(p)
	parts := strings.Split(p, "/")
	for i, seg := range parts {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			parts[i] = "{}"
		}
	}
	return strings.Join(parts, "/")
}

// pathMatch checks whether specPath (from the OpenAPI document) matches
// filterPath (from --includes/--excludes). It supports:
//  1. Trailing-slash normalization (/api/v2/login/ == /api/v2/login)
//  2. Variable-name independence (/api/v2/scans/{scan_id} == /api/v2/scans/{id})
//  3. Method-scoped matching (GET /api/v2/login)
func pathMatch(specPath, filterPath, method string) bool {
	specNorm := normalizePathParam(specPath)
	filterNorm := normalizePathParam(filterPath)

	// Check for method-scoped filter: "get /api/v2/login"
	// We also handle the filterPath containing "#" as separator
	if strings.Contains(filterPath, "#") {
		parts := strings.SplitN(filterPath, "#", 2)
		filterMethod := strings.TrimSpace(strings.ToLower(parts[0]))
		filterPathPart := normalizePathParam(parts[1])
		return normalizePath(method) == filterMethod && specNorm == filterPathPart
	}

	// Check if filter contains a space indicating "METHOD /path"
	if idx := strings.Index(filterPath, " "); idx > 0 {
		filterMethod := strings.TrimSpace(strings.ToLower(filterPath[:idx]))
		filterPathPart := normalizePathParam(strings.TrimSpace(filterPath[idx+1:]))
		return normalizePath(method) == filterMethod && specNorm == filterPathPart
	}

	// Simple path match (ignoring variable names)
	return specNorm == filterNorm
}

func (c *Converter) shouldIncludePath(path, method string) bool {
	// Check excludes first
	for excluded := range c.options.ExcludePaths {
		if pathMatch(path, excluded, method) {
			return false
		}
	}

	// Check includes
	hasIncludes := len(c.options.IncludePaths) > 0
	if !hasIncludes {
		return true
	}

	for included := range c.options.IncludePaths {
		if pathMatch(path, included, method) {
			return true
		}
	}

	return false
}
