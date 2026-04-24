package converter

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
)

const maxToolNameLength = 125

// truncateToolName ensures tool names fit within MCP limits while remaining
// unique Go identifiers. Converts dash/underscore separated names to
// PascalCase (e.g. get_user_list → GetUserList) to save characters and
// improve readability. If still too long, truncates and appends a hash suffix.
func truncateToolName(name string) string {
	if name == toPascalCase(name) && len(name) <= maxToolNameLength {
		return name
	}

	converted := toPascalCase(name)
	if len(converted) <= maxToolNameLength {
		return converted
	}

	h := sha256.Sum256([]byte(name))
	hashStr := fmt.Sprintf("%x", h[:4])
	maxPrefix := maxToolNameLength - len(hashStr) - 1

	var truncated []rune
	for _, r := range converted {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			truncated = append(truncated, r)
		}
		if len(truncated) >= maxPrefix {
			break
		}
	}

	result := string(truncated) + "_" + hashStr
	if len(result) > maxToolNameLength {
		result = result[:maxToolNameLength]
	}

	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		return "_" + result[:maxToolNameLength-1]
	}

	return result
}

// toPascalCase splits by non-alphanumeric separators, lowercases each segment,
// and capitalises the first letter to produce a Go-style identifier.
func toPascalCase(s string) string {
	var b strings.Builder
	capitalizeNext := true
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if capitalizeNext {
				b.WriteRune(unicode.ToUpper(r))
				capitalizeNext = false
			} else {
				b.WriteRune(unicode.ToLower(r))
			}
		} else {
			capitalizeNext = true
		}
	}
	return b.String()
}

// getOperations returns a map of HTTP method to operation
func getOperations(pathItem *openapi3.PathItem) map[string]*openapi3.Operation {
	operations := make(map[string]*openapi3.Operation)

	if pathItem.Get != nil {
		operations["get"] = pathItem.Get
	}
	if pathItem.Post != nil {
		operations["post"] = pathItem.Post
	}
	if pathItem.Put != nil {
		operations["put"] = pathItem.Put
	}
	if pathItem.Delete != nil {
		operations["delete"] = pathItem.Delete
	}
	if pathItem.Options != nil {
		operations["options"] = pathItem.Options
	}
	if pathItem.Head != nil {
		operations["head"] = pathItem.Head
	}
	if pathItem.Patch != nil {
		operations["patch"] = pathItem.Patch
	}
	if pathItem.Trace != nil {
		operations["trace"] = pathItem.Trace
	}

	return operations
}

// convertOperation converts an OpenAPI operation to an MCP tool
func (c *Converter) convertOperation(path, method string, operation *openapi3.Operation) (*Tool, error) {
	// Generate a tool name
	toolName := truncateToolName(c.parser.GetOperationID(path, method, operation))

	// Create the tool
	tool := &Tool{
		Name:        toolName,
		Description: getDescription(operation),
		Args:        []Arg{},
	}

	// Convert parameters to arguments
	args, err := c.convertParameters(operation.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to convert parameters: %w", err)
	}
	if len(args) > 0 {
		tool.Args = append(tool.Args, args...)
	}

	// Convert request body to arguments
	bodyArgs, err := c.convertRequestBody(operation.RequestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request body: %w", err)
	}
	if bodyArgs != nil {
		tool.Args = append(tool.Args, *bodyArgs)
	}

	rawInputSchema, err := GenerateJSONSchemaDraft7(tool.Args)
	if err != nil {
		return nil, fmt.Errorf("failed creating raw input schema for the %s tool input", toolName)
	}

	tool.RawInputSchema = rawInputSchema

	// Sort arguments by name for consistent output
	sort.Slice(tool.Args, func(i, j int) bool {
		return tool.Args[i].Name < tool.Args[j].Name
	})

	// Create request template
	requestTemplate, err := c.createRequestTemplate(path, method, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to create request template: %w", err)
	}
	tool.RequestTemplate = *requestTemplate

	// Create response template
	responseTemplate, err := c.createResponseTemplates(operation)
	if err != nil {
		return nil, fmt.Errorf("failed to create response template: %w", err)
	}
	tool.Responses = responseTemplate

	return tool, nil
}
