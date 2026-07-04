package converter

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Parser represents an OpenAPI parser
type Parser struct {
	doc              *openapi3.T
	ValidateDocument bool
}

// NewParser creates a new OpenAPI parser
func NewParser(validation bool) *Parser {
	return &Parser{
		ValidateDocument: validation, // Default to no validation because there are the parser does not handle OpenAPI 3.1
	}
}

// ParseFile parses an OpenAPI document from a file
func (p *Parser) ParseFile(filePath string) error {
	// Read file first, normalize OAS 3.1, then load via kin-openapi
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	normalized, err := preprocessSpec(data)
	if err != nil {
		return fmt.Errorf("failed to preprocess spec: %w", err)
	}

	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromData(normalized)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI document: %w", err)
	}

	// Validate if needed
	if p.ValidateDocument {
		err = doc.Validate(context.Background())
		if err != nil {
			return fmt.Errorf("invalid OpenAPI document: %w", err)
		}
	}

	p.doc = doc
	return nil
}

// Parse parses an OpenAPI document from bytes
func (p *Parser) Parse(data []byte) error {
	loader := openapi3.NewLoader()

	// Normalize OAS 3.1 features to 3.0-compatible format before loading
	normalized, err := preprocessSpec(data)
	if err != nil {
		return fmt.Errorf("failed to preprocess spec: %w", err)
	}

	// Parse the document (loader can handle both JSON and YAML)
	doc, err := loader.LoadFromData(normalized)

	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI document: %w", err)
	}

	// Validate the document if validation is enabled
	if p.ValidateDocument {
		err = doc.Validate(context.Background())
		if err != nil {
			return fmt.Errorf("invalid OpenAPI document: %w", err)
		}
	}

	p.doc = doc
	return nil
}

// GetDocument returns the parsed OpenAPI document
func (p *Parser) GetDocument() *openapi3.T {
	return p.doc
}

// GetPaths returns all paths in the OpenAPI document
func (p *Parser) GetPaths() map[string]*openapi3.PathItem {
	if p.doc == nil {
		return nil
	}
	return p.doc.Paths.Map()
}

// GetServers returns all servers in the OpenAPI document
func (p *Parser) GetServers() []*openapi3.Server {
	if p.doc == nil {
		return nil
	}
	return p.doc.Servers
}

// GetInfo returns the info section of the OpenAPI document
func (p *Parser) GetInfo() *openapi3.Info {
	if p.doc == nil {
		return nil
	}
	return p.doc.Info
}

// GetOperationID generates an operation ID if one is not provided
func (p *Parser) GetOperationID(path string, method string, operation *openapi3.Operation) string {
	if operation.OperationID != "" {
		return operation.OperationID
	}

	// Generate an operation ID based on the path and method
	pathName := strings.ReplaceAll(path, "/", "_")
	pathName = strings.ReplaceAll(pathName, "{", "")
	pathName = strings.ReplaceAll(pathName, "}", "")
	return fmt.Sprintf("%s%s", strings.ToLower(method), pathName)
}
