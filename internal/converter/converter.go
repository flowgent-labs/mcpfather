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
		p = strings.TrimSpace(p)
		if p != "" {
			includeSet[p] = struct{}{}
		}
	}
	for _, p := range excludePaths {
		p = strings.TrimSpace(p)
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

func (c *Converter) shouldIncludePath(path, method string) bool {
	pathLower := strings.ToLower(path)
	methodLower := strings.ToLower(method)

	// Check excludes first
	for excluded := range c.options.ExcludePaths {
		excludedLower := strings.ToLower(excluded)
		if pathLower == excludedLower {
			return false
		}
		if pathLower+"#"+methodLower == excludedLower {
			return false
		}
	}

	// Check includes
	hasIncludes := len(c.options.IncludePaths) > 0
	if !hasIncludes {
		return true
	}

	for included := range c.options.IncludePaths {
		includedLower := strings.ToLower(included)
		if pathLower == includedLower {
			return true
		}
		if pathLower+"#"+methodLower == includedLower {
			return true
		}
	}

	return false
}
