package converter

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
)

// Converter represents an OpenAPI to MCP converter
type Converter struct {
	parser  *Parser
	options ConvertOptions
	verbose bool
}

type ConverterInterface interface {
	Convert() (*MCPConfig, error)
}

// NewConverter creates a new OpenAPI to MCP converter
func NewConverter(parser *Parser, includePaths []string, excludePaths []string, verbose bool) (*Converter, error) {
	includeSet := make(map[string]struct{})
	excludeSet := make(map[string]struct{})

	for _, p := range includePaths {
		cleaned := cleanFilterPath(p)
		if cleaned != "" {
			includeSet[cleaned] = struct{}{}
		}
		if verbose && cleaned != "" && cleaned != cleanFilterPath(p) {
			fmt.Fprintf(os.Stderr, "[verbose] include path cleaned: %q -> %q\n", p, cleaned)
		}
	}
	for _, p := range excludePaths {
		cleaned := cleanFilterPath(p)
		if cleaned != "" {
			excludeSet[cleaned] = struct{}{}
		}
		if verbose && cleaned != "" && cleaned != cleanFilterPath(p) {
			fmt.Fprintf(os.Stderr, "[verbose] exclude path cleaned: %q -> %q\n", p, cleaned)
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
		verbose: verbose,
	}, nil
}


// Convert converts an OpenAPI document to an MCP configuration
func (c *Converter) Convert() (*MCPConfig, error) {
	if c.parser.GetDocument() == nil {
		return nil, fmt.Errorf("no OpenAPI document loaded")
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "[verbose] converting OpenAPI document with %d paths\n", len(c.parser.GetPaths()))
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
				if c.verbose {
					operationID := c.parser.GetOperationID(path, method, operation)
					fmt.Fprintf(os.Stderr, "[verbose] filtered out: %s %s (operationId=%s)\n", method, path, operationID)
				}
				continue
			}

			if c.verbose {
				operationID := c.parser.GetOperationID(path, method, operation)
				fmt.Fprintf(os.Stderr, "[verbose] including: %s %s (operationId=%s)\n", method, path, operationID)
			}

			tool, err := c.convertOperation(path, method, operation)
			if err != nil {
				return nil, fmt.Errorf("failed to convert operation %s %s: %w", method, path, err)
			}
			if c.verbose {
				fmt.Fprintf(os.Stderr, "[verbose] tool created: %s\n", tool.Name)
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

// cleanFilterPath aggressively strips all invisible, control, and
// formatting characters from a filter path so that values pasted
// from any source (PowerShell, bash, YAML, browsers, terminal
// emulators, Windows editors) still match the OpenAPI spec paths.
func cleanFilterPath(p string) string {
	// Build a new string with only meaningful characters.
	var b strings.Builder
	for _, r := range p {
		switch r {
		case '\t', '\n', '\r', '\v', '\f':
			// skip all whitespace/control chars
		default:
			// skip any unicode space or control category
			if !unicode.IsSpace(r) && !unicode.IsControl(r) {
				b.WriteRune(r)
			}
		}
	}
	p = b.String()
	// Remove surrounding quotes (PowerShell/bash)
	if len(p) >= 2 && ((p[0] == '"' && p[len(p)-1] == '"') || (p[0] == '\'' && p[len(p)-1] == '\'')) {
		p = p[1 : len(p)-1]
	}
	// Strip leading/trailing slashes and colons (YAML path separator)
	p = strings.Trim(p, "/:")
	// Remove any embedded \r\n, \n, \r sequences that may have
	// been introduced by line-wrapped paste operations
	p = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return -1
		}
		return r
	}, p)
	if p == "" {
		return "/"
	}
	return p
}

// normalizePath aggressively strips ALL invisible, control, and
// formatting characters, leading/trailing slashes, and lowercases
// for consistent matching across all platforms.
func normalizePath(p string) string {
	// Remove all control chars, whitespace, BOM, etc.
	p = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || unicode.IsControl(r) || r == '\uFEFF' {
			return -1
		}
		return r
	}, p)
	p = strings.Trim(p, "/:")
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
//  1. Trailing-slash/colon normalization (/api/v2/login/ == /api/v2/login)
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

	// Exact path match (ignoring variable names, trailing slashes, colons)
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
