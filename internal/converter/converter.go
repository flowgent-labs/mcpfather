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
		cleaned := cleanOperationId(p)
		if cleaned != "" {
			includeSet[cleaned] = struct{}{}
		}
	}
	for _, p := range excludePaths {
		cleaned := cleanOperationId(p)
		if cleaned != "" {
			excludeSet[cleaned] = struct{}{}
		}
	}

	// Check for conflicts
	for p := range includeSet {
		if _, ok := excludeSet[p]; ok {
			return nil, fmt.Errorf("operationId '%s' is specified in both --includes and --excludes", p)
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
			operationID := c.parser.GetOperationID(path, method, operation)
			if !c.shouldIncludeOperation(operationID) {
				if c.verbose {
					fmt.Fprintf(os.Stderr, "[verbose] filtered out: %s %s (operationId=%s)\n", method, path, operationID)
				}
				continue
			}

			if c.verbose {
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

// cleanOperationId strips invisible/control characters and quotes from
// an operationId filter value so that values pasted from any source
// (PowerShell, bash, terminal, browser) still match the spec exactly.
func cleanOperationId(p string) string {
	var b strings.Builder
	for _, r := range p {
		if !unicode.IsSpace(r) && !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	p = b.String()
	// Remove surrounding quotes
	if len(p) >= 2 && ((p[0] == '"' && p[len(p)-1] == '"') || (p[0] == '\'' && p[len(p)-1] == '\'')) {
		p = p[1 : len(p)-1]
	}
	return p
}

func (c *Converter) shouldIncludeOperation(operationID string) bool {
	// Check excludes first
	for excluded := range c.options.ExcludePaths {
		if operationID == excluded {
			return false
		}
	}

	// Check includes
	hasIncludes := len(c.options.IncludePaths) > 0
	if !hasIncludes {
		return true
	}

	for included := range c.options.IncludePaths {
		if operationID == included {
			return true
		}
	}

	return false
}
