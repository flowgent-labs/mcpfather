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

	// Detect upload-type APIs. An operation may define multiple request body
	// media types. Use explicit media type selection rather than inferring
	// multipart from arbitrary JSON schemas: multipart/form-data wins when
	// present, otherwise raw binary media types use the single FileRef path.
	if ct := detectUploadContentType(operation); ct != "" {
		if ct == "multipart/form-data" {
			fileArgs, formArgs := c.extractMultipartFileArgs(operation)
			if len(fileArgs) > 0 {
				replaceBodyWithFileRefArgs(tool, fileArgs, formArgs)
			} else if isMultipartPlainBinary(operation) {
				// Non-object multipart schema (e.g. "type: string, format: binary"):
				// treat as a single FileArg — unified file-ref approach (no base64).
				fileArgs := []FileArg{{Name: "file", Description: "File to upload", Required: true}}
				replaceBodyWithFileRefArgs(tool, fileArgs, nil)
			}
			// Object schema without binary properties: fall through to
			// standard JSON body processing (no UploadContentType set).
		} else {
			// Non-multipart upload (application/octet-stream, image/*,
			// video/*, audio/*): unified file-ref approach — single FileArg
			// with URI reference. The handler downloads the file and
			// forwards it as raw binary body with the original content type.
			fileArgs := []FileArg{{Name: "file", Description: "File to upload", Required: true}}
			replaceBodyWithFileRefArgs(tool, fileArgs, nil)
			tool.UploadContentType = ct
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
	contentTypes := sortedContentTypes(operation.RequestBody.Value.Content)
	for _, ct := range contentTypes {
		if isMultipartContentType(ct) {
			return "multipart/form-data"
		}
	}
	for _, ct := range contentTypes {
		baseCT := baseContentType(ct)
		if baseCT == "application/octet-stream" ||
			strings.HasPrefix(baseCT, "image/") || strings.HasPrefix(baseCT, "video/") ||
			strings.HasPrefix(baseCT, "audio/") {
			return ct
		}
	}
	return ""
}

func baseContentType(ct string) string {
	base, _, _ := strings.Cut(ct, ";")
	return strings.ToLower(strings.TrimSpace(base))
}

func isMultipartContentType(ct string) bool {
	return baseContentType(ct) == "multipart/form-data"
}

// extractMultipartFileArgs parses multipart/form-data object schemas to extract
// file-type properties as FileArgs and non-file properties as regular Args.
// Binary file properties may be direct fields or hidden behind anyOf/oneOf/allOf
// wrappers, but only explicit multipart/form-data request bodies take this path.
func (c *Converter) extractMultipartFileArgs(operation *openapi3.Operation) ([]FileArg, []Arg) {
	if operation == nil || operation.RequestBody == nil || operation.RequestBody.Value == nil {
		return nil, nil
	}

	for _, ct := range sortedContentTypes(operation.RequestBody.Value.Content) {
		if !isMultipartContentType(ct) {
			continue
		}
		mediaType := operation.RequestBody.Value.Content[ct]
		if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
			continue
		}
		fileArgs, formArgs := c.extractMultipartFileArgsFromSchema(mediaType.Schema.Value, multipartEncodingContentTypes(mediaType))
		if len(fileArgs) > 0 {
			return fileArgs, formArgs
		}
	}
	return nil, nil
}

func multipartEncodingContentTypes(mediaType *openapi3.MediaType) map[string]string {
	if mediaType == nil || len(mediaType.Encoding) == 0 {
		return nil
	}
	out := make(map[string]string, len(mediaType.Encoding))
	for name, enc := range mediaType.Encoding {
		if enc == nil {
			continue
		}
		if ct := strings.TrimSpace(enc.ContentType); ct != "" {
			out[name] = ct
		}
	}
	return out
}

func replaceBodyWithFileRefArgs(tool *Tool, fileArgs []FileArg, formArgs []Arg) {
	filteredArgs := make([]Arg, 0, len(tool.Args)+len(fileArgs)+len(formArgs))
	for _, a := range tool.Args {
		if a.Name != "body" {
			filteredArgs = append(filteredArgs, a)
		}
	}
	filteredArgs = append(filteredArgs, formArgs...)
	for _, fa := range fileArgs {
		filteredArgs = append(filteredArgs, Arg{
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
	tool.Args = filteredArgs
	tool.FileArgs = fileArgs
}

func (c *Converter) extractMultipartFileArgsFromSchema(schema *openapi3.Schema, partContentTypes map[string]string) ([]FileArg, []Arg) {
	properties := make(map[string]*openapi3.Schema)
	requiredSet := make(map[string]bool)
	collectObjectProperties(schema, properties, requiredSet, true, make(map[*openapi3.Schema]bool))
	if len(properties) == 0 {
		return nil, nil
	}

	propNames := make([]string, 0, len(properties))
	for name := range properties {
		propNames = append(propNames, name)
	}
	sort.Strings(propNames)

	var fileArgs []FileArg
	var formArgs []Arg
	for _, name := range propNames {
		prop := properties[name]
		if prop == nil {
			continue
		}
		if isBinaryFileSchema(prop, make(map[*openapi3.Schema]bool)) {
			fileArgs = append(fileArgs, FileArg{
				Name:        name,
				Description: prop.Description,
				Required:    requiredSet[name],
				ContentType: strings.TrimSpace(partContentTypes[name]),
			})
			continue
		}

		propSchema, err := c.applySchema(prop, make(map[*openapi3.Schema]bool))
		if err != nil {
			continue
		}
		formArg := Arg{
			Name:                 name,
			Description:          prop.Description,
			Source:               "body",
			Required:             requiredSet[name],
			MultipartContentType: strings.TrimSpace(partContentTypes[name]),
		}
		if propSchema != nil {
			formArg.Schema = propSchema
		}
		formArgs = append(formArgs, formArg)
	}
	return fileArgs, formArgs
}

func collectObjectProperties(schema *openapi3.Schema, properties map[string]*openapi3.Schema, required map[string]bool, requiredApplies bool, visited map[*openapi3.Schema]bool) {
	if schema == nil || visited[schema] {
		return
	}
	visited[schema] = true

	if len(schema.Properties) > 0 {
		if requiredApplies {
			for _, name := range schema.Required {
				required[name] = true
			}
		}
		for name, ref := range schema.Properties {
			if ref != nil && ref.Value != nil {
				properties[name] = ref.Value
			}
		}
	}

	for _, ref := range schema.AllOf {
		if ref != nil && ref.Value != nil {
			collectObjectProperties(ref.Value, properties, required, requiredApplies, visited)
		}
	}
	for _, ref := range schema.OneOf {
		if ref != nil && ref.Value != nil {
			collectObjectProperties(ref.Value, properties, required, false, visited)
		}
	}
	for _, ref := range schema.AnyOf {
		if ref != nil && ref.Value != nil {
			collectObjectProperties(ref.Value, properties, required, false, visited)
		}
	}
}

func isBinaryFileSchema(schema *openapi3.Schema, visited map[*openapi3.Schema]bool) bool {
	if schema == nil || visited[schema] {
		return false
	}
	visited[schema] = true

	if schema.Format == "binary" && (schema.Type == nil || contains(*schema.Type, "string")) {
		return true
	}
	if hasArrayType(schema) && schema.Items != nil && schema.Items.Value != nil {
		if isBinaryFileSchema(schema.Items.Value, visited) {
			return true
		}
	}
	for _, ref := range schema.AllOf {
		if ref != nil && ref.Value != nil && isBinaryFileSchema(ref.Value, visited) {
			return true
		}
	}
	for _, ref := range schema.OneOf {
		if ref != nil && ref.Value != nil && isBinaryFileSchema(ref.Value, visited) {
			return true
		}
	}
	for _, ref := range schema.AnyOf {
		if ref != nil && ref.Value != nil && isBinaryFileSchema(ref.Value, visited) {
			return true
		}
	}
	return false
}

// isMultipartPlainBinary returns true when the operation's multipart/form-data
// schema is a plain binary type (not an object with named properties). This
// case uses the unified file-ref approach with a single "file" FileArg.
func isMultipartPlainBinary(operation *openapi3.Operation) bool {
	if operation == nil || operation.RequestBody == nil || operation.RequestBody.Value == nil {
		return false
	}
	for ct, mediaType := range operation.RequestBody.Value.Content {
		if !isMultipartContentType(ct) {
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
