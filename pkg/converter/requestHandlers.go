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

// toPascalCase splits on non-alphanumeric separators and camelCase boundaries,
// lowercasing each segment and capitalising the first letter to produce a Go identifier.
func toPascalCase(s string) string {
	var b strings.Builder
	capitalizeNext := true
	var prev rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if capitalizeNext {
				b.WriteRune(unicode.ToUpper(r))
				capitalizeNext = false
			} else if unicode.IsUpper(r) && unicode.IsLower(prev) {
				// camelCase boundary: lowercase→uppercase transition marks a new word
				b.WriteRune(r)
			} else {
				b.WriteRune(unicode.ToLower(r))
			}
			prev = r
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
	operationID := c.parser.GetOperationID(path, method, operation)
	toolName := truncateToolName(operationID)

	// Create the tool
	tool := &Tool{
		Name:        toolName,
		OperationID: operationID,
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

	// Detect upload-type API (request body expects file).
	if ct := detectUploadContentType(operation); ct != "" {
		if ct == "multipart/form-data" {
			// For multipart/form-data, prefer the FileRef approach: binary file
			// properties are converted to URI (string) arguments. The generated
			// handler download each URI to a local temp directory and builds a
			// proper multipart/form-data request to the upstream. This avoids
			// the token/memory explosion of base64-encoding file content.
			fileArgs, formArgs := c.extractMultipartFileArgs(operation)
			if len(fileArgs) > 0 {
				// Replace the single body arg (from convertRequestBody) with
				// individual form field args + file ref args.
				filteredArgs := make([]Arg, 0, len(tool.Args))
				for _, a := range tool.Args {
					if a.Name != "body" {
						filteredArgs = append(filteredArgs, a)
					}
				}
				tool.Args = filteredArgs

				// Add non-file form field args
				tool.Args = append(tool.Args, formArgs...)

				// Add file ref args (URI type)
				for _, fa := range fileArgs {
					tool.Args = append(tool.Args, Arg{
						Name:        fa.Name,
						Description: fa.Description + " (URI reference — the file will be downloaded and uploaded to the upstream service)",
						Source:      "body",
						Required:    fa.Required,
						Schema: &Schema{
							Types:  []string{"string"},
							Format: "uri",
						},
					})
				}
				tool.FileArgs = fileArgs
			} else if isMultipartPlainBinary(operation) {
				// Non-object multipart schema (e.g. "type: string, format: binary"):
				// use the legacy file_name / file_content upload path.
				tool.UploadContentType = ct
				tool.Args = append(tool.Args,
					Arg{
						Name:        "file_name",
						Description: "File name to upload (looked up in ~/.{project}/upload/; also used when file_content is staged). When omitted, the tool sends a standard JSON request body instead.",
						Source:      "body",
						Required:    false,
					},
					Arg{
						Name:        "file_content",
						Description: "Base64-encoded file content (use this in HTTP mode to send file data inline)",
						Source:      "body",
						Required:    false,
					},
				)
				for i := range tool.Args {
					if tool.Args[i].Name == "body" {
						tool.Args[i].Required = false
						break
					}
				}
			}
			// Object schema without binary properties: fall through to
			// standard JSON body processing (no UploadContentType set).
		} else {
			// Non-multipart upload (application/octet-stream, image/*,
			// video/*, audio/*): use file_name / file_content approach.
			tool.UploadContentType = ct
			// Upload tools accept file_name and file_content (both optional).
			// - file_name: looked up in ~/.{project}/upload/ (stdio mode) or used when
			//   staging file_content (HTTP mode).
			// - file_content: base64-encoded file data (HTTP mode).
			// When neither is provided, the tool falls back to standard JSON body.
			tool.Args = append(tool.Args,
				Arg{
					Name:        "file_name",
					Description: "File name to upload (looked up in ~/.{project}/upload/; also used when file_content is staged). When omitted, the tool sends a standard JSON request body instead.",
					Source:      "body",
					Required:    false,
				},
				Arg{
					Name:        "file_content",
					Description: "Base64-encoded file content (use this in HTTP mode to send file data inline)",
					Source:      "body",
					Required:    false,
				},
			)
			// When upload content type is set, the original body arg (from
			// convertRequestBody) should also be optional so that tools
			// work without requiring a file or body.
			for i := range tool.Args {
				if tool.Args[i].Name == "body" {
					tool.Args[i].Required = false
					break
				}
			}
		}
	}

	// Sort arguments by name for consistent output
	sort.Slice(tool.Args, func(i, j int) bool {
		return tool.Args[i].Name < tool.Args[j].Name
	})

	// Generate input schema after all args (including upload local_file_path) are added
	rawInputSchema, err := GenerateJSONSchemaDraft7(tool.Args)
	if err != nil {
		return nil, fmt.Errorf("failed creating raw input schema for the %s tool input", toolName)
	}

	tool.RawInputSchema = rawInputSchema

	// Sort arguments by name for consistent output
	sort.Slice(tool.Args, func(i, j int) bool {
		return tool.Args[i].Name < tool.Args[j].Name
	})

	return tool, nil
}

// detectUploadContentType returns the content type if the operation expects a
// file upload body (multipart/form-data, application/octet-stream, image/*, video/*, audio/*),
// or "" otherwise.
func detectUploadContentType(operation *openapi3.Operation) string {
	if operation == nil || operation.RequestBody == nil || operation.RequestBody.Value == nil {
		return ""
	}
	for ct := range operation.RequestBody.Value.Content {
		if ct == "multipart/form-data" ||
			ct == "application/octet-stream" ||
			strings.HasPrefix(ct, "image/") || strings.HasPrefix(ct, "video/") ||
			strings.HasPrefix(ct, "audio/") {
			return ct
		}
	}
	return ""
}

// extractMultipartFileArgs parses a multipart/form-data request body schema to
// extract file-type properties (format: binary) as FileArgs and non-file
// properties as regular Args. Returns nil slices when the schema is not an
// object with properties (e.g., a plain binary string schema without named
// fields cannot be decomposed into individual FileRef args).
func (c *Converter) extractMultipartFileArgs(operation *openapi3.Operation) ([]FileArg, []Arg) {
	if operation == nil || operation.RequestBody == nil || operation.RequestBody.Value == nil {
		return nil, nil
	}

	for ct, mediaType := range operation.RequestBody.Value.Content {
		if ct != "multipart/form-data" {
			continue
		}
		if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
			continue
		}
		schema := mediaType.Schema.Value
		// Only handle object schemas with named properties.
		// A plain "type: string, format: binary" multipart body is handled by
		// the legacy file_name/file_content path (UploadContentType).
		if schema.Type == nil || !contains(*schema.Type, "object") || schema.Properties == nil {
			return nil, nil
		}

		var fileArgs []FileArg
		var formArgs []Arg
		requiredSet := make(map[string]bool)
		for _, r := range schema.Required {
			requiredSet[r] = true
		}

		visited := make(map[*openapi3.Schema]bool)
		for name, propRef := range schema.Properties {
			if propRef == nil || propRef.Value == nil {
				continue
			}
			prop := propRef.Value
			if prop.Format == "binary" {
				fileArgs = append(fileArgs, FileArg{
					Name:        name,
					Description: prop.Description,
					Required:    requiredSet[name],
				})
			} else {
				// Non-file form field — convert to a regular body arg.
				propSchema, err := c.applySchema(prop, visited)
				if err != nil {
					continue
				}
				formArg := Arg{
					Name:        name,
					Description: prop.Description,
					Source:      "body",
					Required:    requiredSet[name],
				}
				if propSchema != nil {
					formArg.Schema = propSchema
				}
				formArgs = append(formArgs, formArg)
			}
		}
		return fileArgs, formArgs
	}
	return nil, nil
}

// isMultipartPlainBinary returns true when the operation's multipart/form-data
// schema is a plain binary type (not an object with named properties). This
// case cannot use FileRef (there are no named file args to extract) and must
// fall back to the legacy file_name / file_content approach.
func isMultipartPlainBinary(operation *openapi3.Operation) bool {
	if operation == nil || operation.RequestBody == nil || operation.RequestBody.Value == nil {
		return false
	}
	for ct, mediaType := range operation.RequestBody.Value.Content {
		if ct != "multipart/form-data" {
			continue
		}
		if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
			continue
		}
		schema := mediaType.Schema.Value
		// Plain binary: not an object schema, but has format: binary
		return schema.Type != nil &&
			!contains(*schema.Type, "object") &&
			schema.Format == "binary"
	}
	return false
}
