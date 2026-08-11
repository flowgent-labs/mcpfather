package converter

import (
	"os"
	"testing"
)

// testSpecOAS30 is a minimal OpenAPI 3.0 spec used by unit tests in this package.
const testSpecOAS30 = `openapi: "3.0.3"
info:
  title: Blogs API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /posts:
    get:
      operationId: listPosts
      summary: List all blog posts
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PostList'
  /posts/{id}:
    get:
      operationId: getPost
      summary: Get a blog post by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        "200":
          description: OK
    delete:
      operationId: deletePost
      summary: Delete a blog post
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        "204":
          description: OK
  /attachments:
    post:
      operationId: uploadAttachment
      summary: Upload an attachment
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                file:
                  type: string
                  format: binary
      responses:
        "200":
          description: OK
  /attachments/{id}:
    get:
      operationId: downloadAttachment
      summary: Download an attachment
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        "200":
          description: OK
          content:
            application/octet-stream:
              schema:
                type: string
                format: binary
  /search:
    get:
      operationId: searchPosts
      summary: Search blog posts
      parameters:
        - name: q
          in: query
          schema:
            type: string
      responses:
        "200":
          description: OK
components:
  schemas:
    PostList:
      type: object
      properties:
        posts:
          type: array
          items:
            type: object
            properties:
              id:
                type: integer
              title:
                  type: string
`

// writeTestSpecOAS30 writes the test spec to a temp file and returns its path.
func writeTestSpecOAS30(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "testspec-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp spec file: %v", err)
	}
	if _, err := f.WriteString(testSpecOAS30); err != nil {
		f.Close()
		t.Fatalf("Failed to write temp spec file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close temp spec file: %v", err)
	}
	return f.Name()
}

func TestNewConverter(t *testing.T) {
	parser := NewParser(false)
	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil Converter")
	}
	if c.parser != parser {
		t.Error("expected parser to be set")
	}
	if c.options.ServerConfig == nil {
		t.Error("expected ServerConfig to be initialized")
	}
}

func TestConverter_Convert(t *testing.T) {
	parser := NewParser(false)
	if err := parser.Parse([]byte(testSpecOAS30)); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if config == nil {
		t.Fatal("expected non-nil MCPConfig")
	}
	if config.Server.Config == nil {
		t.Error("expected Server.Config to be set")
	}
	if len(config.Tools) == 0 {
		t.Error("expected at least one tool in Tools")
	}
	// Check that tools are sorted by name
	for i := 1; i < len(config.Tools); i++ {
		if config.Tools[i-1].Name > config.Tools[i].Name {
			t.Errorf("tools not sorted by name: %q > %q", config.Tools[i-1].Name, config.Tools[i].Name)
		}
	}
}

func TestCleanOperationId(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"listSpaces", "listSpaces"},
		{"'listSpaces'", "listSpaces"},
		{`"listSpaces"`, "listSpaces"},
		{"  listSpaces  ", "listSpaces"},
		{"listSpaces\n", "listSpaces"},
		{"listSpaces\r\n", "listSpaces"},
		{"get-a-very-long-operation-id", "get-a-very-long-operation-id"},
		{"", ""},
		{"''", ""},
		{`""`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanOperationId(tt.input)
			if got != tt.want {
				t.Errorf("cleanOperationId(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConverter_Convert_IncludeExcludeByOperationId(t *testing.T) {
	parser := NewParser(false)
	if err := parser.Parse([]byte(testSpecOAS30)); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	// Include only "listPosts"
	c, err := NewConverter(parser, []string{"listPosts"}, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}
	if len(config.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(config.Tools))
	}
	if config.Tools[0].Name != "ListPosts" {
		t.Errorf("expected tool ListPosts, got %s", config.Tools[0].Name)
	}
}

func TestConverter_Convert_NoDocument(t *testing.T) {
	parser := NewParser(false)
	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	_, err = c.Convert()
	if err == nil {
		t.Fatal("expected error when no OpenAPI document is loaded")
	}
}

func TestConverter_UploadDownloadDetection(t *testing.T) {
	parser := NewParser(false)
	if err := parser.Parse([]byte(testSpecOAS30)); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Find upload and download tools
	var uploadTool *Tool
	var downloadTool *Tool
	for i := range config.Tools {
		switch config.Tools[i].Name {
		case "UploadAttachment":
			uploadTool = &config.Tools[i]
		case "DownloadAttachment":
			downloadTool = &config.Tools[i]
		}
	}

	if uploadTool == nil {
		t.Fatal("expected uploadAttachment tool to be generated")
	}
	// With FileRef approach, multipart/form-data object schemas use FileArgs
	// instead of UploadContentType.
	if len(uploadTool.FileArgs) == 0 {
		t.Error("expected uploadAttachment to have FileArgs set (FileRef for multipart/form-data)")
	}
	if len(uploadTool.FileArgs) > 0 && uploadTool.FileArgs[0].Name != "file" {
		t.Errorf("expected FileArgs[0].Name = 'file', got %q", uploadTool.FileArgs[0].Name)
	}
	if uploadTool.UploadContentType != "" {
		t.Error("multipart/form-data with object schema should use FileArgs, not UploadContentType")
	}

	if downloadTool == nil {
		t.Fatal("expected downloadAttachment tool to be generated")
	}
	if downloadTool.UploadContentType != "" {
		t.Error("download tool should not have UploadContentType set")
	}
	if len(downloadTool.FileArgs) != 0 {
		t.Error("download tool should not have FileArgs set")
	}
}

func TestConverter_UploadTool_FileRefArgs(t *testing.T) {
	parser := NewParser(false)
	if err := parser.Parse([]byte(testSpecOAS30)); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Find upload tool
	var uploadTool *Tool
	for i := range config.Tools {
		if config.Tools[i].Name == "UploadAttachment" {
			uploadTool = &config.Tools[i]
			break
		}
	}
	if uploadTool == nil {
		t.Fatal("expected uploadAttachment tool to be generated")
	}

	// Verify FileArgs is populated (FileRef approach for multipart/form-data)
	if len(uploadTool.FileArgs) == 0 {
		t.Fatal("expected FileArgs to be set for multipart/form-data tool")
	}
	if uploadTool.FileArgs[0].Name != "file" {
		t.Errorf("expected FileArgs[0].Name = 'file', got %q", uploadTool.FileArgs[0].Name)
	}

	// Verify the 'file' arg is uri type (FileRef)
	var fileArg *Arg
	for i := range uploadTool.Args {
		if uploadTool.Args[i].Name == "file" {
			fileArg = &uploadTool.Args[i]
			break
		}
	}
	if fileArg == nil {
		t.Fatal("expected 'file' uri arg (FileRef) in upload tool")
	}
	if fileArg.Schema == nil || fileArg.Schema.Format != "uri" {
		t.Error("file arg should have uri format (FileRef)")
	}

	// file_name and file_content should NOT be present (replaced by FileRef)
	for _, arg := range uploadTool.Args {
		if arg.Name == "file_name" || arg.Name == "file_content" {
			t.Errorf("FileRef tools should not have %s arg (use uri args instead)", arg.Name)
		}
	}

	// body arg should NOT be present (replaced by individual form + file ref args)
	for _, arg := range uploadTool.Args {
		if arg.Name == "body" {
			t.Error("FileRef tools should not have a 'body' arg")
		}
	}
}

// testSpecFormURLEncoded is a spec with application/x-www-form-urlencoded request body
const testSpecFormURLEncoded = `openapi: "3.0.3"
info:
  title: Form URL Encoded API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /form:
    post:
      operationId: submitForm
      summary: Submit a form
      requestBody:
        required: true
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              properties:
                name:
                  type: string
                email:
                  type: string
      responses:
        "200":
          description: OK
`

func TestDetectUploadContentType_ExcludesFormUrlEncoded(t *testing.T) {
	parser := NewParser(false)
	if err := parser.Parse([]byte(testSpecFormURLEncoded)); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if len(config.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(config.Tools))
	}

	tool := config.Tools[0]
	if tool.UploadContentType != "" {
		t.Errorf("form-url-encoded API should NOT have UploadContentType set, got %q", tool.UploadContentType)
	}

	// Verify no file_name or file_content args are added
	for _, arg := range tool.Args {
		if arg.Name == "file_name" || arg.Name == "file_content" {
			t.Errorf("form-url-encoded API should not have %s arg", arg.Name)
		}
	}
}

// testSpecMultipartMixed is a multipart/form-data spec with both file (binary)
// and non-file (text) properties to exercise the full FileRef extraction.
const testSpecMultipartMixed = `openapi: "3.0.3"
info:
  title: Multipart Mixed API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /resource:
    post:
      operationId: createResource
      summary: Create a resource with file attachment
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              required:
                - name
              properties:
                name:
                  type: string
                  description: Resource name
                description:
                  type: string
                  description: Resource description
                metadata:
                  type: object
                  description: JSON metadata
                  properties:
                    enabled:
                      type: boolean
                attachment:
                  type: string
                  format: binary
                  description: Optional file
                photo:
                  type: string
                  format: binary
                  description: Optional photo
            encoding:
              metadata:
                contentType: application/json
              attachment:
                contentType: application/zip
      responses:
        "200":
          description: OK
`

func TestExtractMultipartFileArgs_MixedProperties(t *testing.T) {
	parser := NewParser(false)
	if err := parser.Parse([]byte(testSpecMultipartMixed)); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if len(config.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(config.Tools))
	}
	tool := config.Tools[0]

	// Should have FileArgs for both binary properties
	if len(tool.FileArgs) != 2 {
		t.Fatalf("expected 2 FileArgs (attachment + photo), got %d", len(tool.FileArgs))
	}
	// Check FileArgs content
	fileNames := make(map[string]bool)
	fileContentTypes := make(map[string]string)
	for _, fa := range tool.FileArgs {
		fileNames[fa.Name] = true
		fileContentTypes[fa.Name] = fa.ContentType
	}
	if !fileNames["attachment"] {
		t.Error("expected FileArg for 'attachment'")
	}
	if got := fileContentTypes["attachment"]; got != "application/zip" {
		t.Errorf("attachment ContentType = %q, want application/zip", got)
	}
	if !fileNames["photo"] {
		t.Error("expected FileArg for 'photo'")
	}

	// No UploadContentType (FileRef takes precedence)
	if tool.UploadContentType != "" {
		t.Error("UploadContentType should be empty for multipart FileRef tools")
	}

	// Check args: should have name (required), description, attachment (uri), photo (uri)
	argNames := make(map[string]*Arg)
	for i := range tool.Args {
		argNames[tool.Args[i].Name] = &tool.Args[i]
	}

	// name should be required string
	if nameArg, ok := argNames["name"]; !ok {
		t.Error("expected 'name' arg (form field)")
	} else if !nameArg.Required {
		t.Error("'name' should be required (schema says required: [name])")
	}

	// description should be optional string
	if descArg, ok := argNames["description"]; !ok {
		t.Error("expected 'description' arg (form field)")
	} else if descArg.Required {
		t.Error("'description' should be optional")
	}

	if metadataArg, ok := argNames["metadata"]; !ok {
		t.Error("expected 'metadata' arg (form field)")
	} else if metadataArg.MultipartContentType != "application/json" {
		t.Errorf("metadata MultipartContentType = %q, want application/json", metadataArg.MultipartContentType)
	}

	// attachment should be optional uri
	if attArg, ok := argNames["attachment"]; !ok {
		t.Error("expected 'attachment' arg (file ref)")
	} else {
		if attArg.Schema == nil || attArg.Schema.Format != "uri" {
			t.Error("attachment should have uri format")
		}
		if attArg.Required {
			t.Error("attachment should be optional")
		}
	}

	// photo should be optional uri
	if photoArg, ok := argNames["photo"]; !ok {
		t.Error("expected 'photo' arg (file ref)")
	} else {
		if photoArg.Schema == nil || photoArg.Schema.Format != "uri" {
			t.Error("photo should have uri format")
		}
	}

	// No body, file_name, or file_content args
	for _, badName := range []string{"body", "file_name", "file_content"} {
		if _, ok := argNames[badName]; ok {
			t.Errorf("tool should NOT have %q arg", badName)
		}
	}
}

const testSpecMultipartBinaryCompatibility = `openapi: "3.0.3"
info:
  title: Multipart Compatibility API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /direct:
    post:
      operationId: uploadReportDirect
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              required:
                - file
              properties:
                file:
                  type: string
                  format: binary
                  description: Report file
                note:
                  type: string
      responses:
        "200":
          description: OK
  /anyof:
    post:
      operationId: uploadReportAnyOf
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                file:
                  anyOf:
                    - type: string
                      format: binary
                    - type: string
                  title: File
                note:
                  type: string
      responses:
        "200":
          description: OK
  /allof:
    post:
      operationId: uploadReportAllOf
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              allOf:
                - type: object
                  required:
                    - file
                  properties:
                    file:
                      oneOf:
                        - type: string
                          format: binary
                        - type: string
                    note:
                      type: string
      responses:
        "200":
          description: OK
  /array:
    post:
      operationId: uploadReportArray
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                file:
                  type: array
                  items:
                    type: string
                    format: binary
      responses:
        "200":
          description: OK
  /multi-content:
    post:
      operationId: uploadReportMultiContent
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                file:
                  type: string
                  format: binary
                note:
                  type: string
          multipart/form-data:
            schema:
              type: object
              properties:
                file:
                  anyOf:
                    - type: string
                      format: binary
                    - type: string
                note:
                  type: string
      responses:
        "200":
          description: OK
`

func TestMultipartBinarySchemasUseFileRefs(t *testing.T) {
	parser := NewParser(false)
	if err := parser.Parse([]byte(testSpecMultipartBinaryCompatibility)); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	tools := make(map[string]Tool)
	for _, tool := range config.Tools {
		tools[tool.OperationID] = tool
	}

	tests := []struct {
		operationID  string
		requiredFile bool
		wantNote     bool
	}{
		{operationID: "uploadReportDirect", requiredFile: true, wantNote: true},
		{operationID: "uploadReportAnyOf", requiredFile: false, wantNote: true},
		{operationID: "uploadReportAllOf", requiredFile: true, wantNote: true},
		{operationID: "uploadReportArray", requiredFile: false, wantNote: false},
		{operationID: "uploadReportMultiContent", requiredFile: false, wantNote: true},
	}

	for _, tt := range tests {
		t.Run(tt.operationID, func(t *testing.T) {
			tool, ok := tools[tt.operationID]
			if !ok {
				t.Fatalf("missing tool for operationID %q", tt.operationID)
			}
			if tool.UploadContentType != "" {
				t.Fatalf("named multipart file fields should not use raw UploadContentType, got %q", tool.UploadContentType)
			}
			if len(tool.FileArgs) != 1 {
				t.Fatalf("expected exactly one FileArg, got %d", len(tool.FileArgs))
			}
			if tool.FileArgs[0].Name != "file" {
				t.Fatalf("FileArg name = %q, want file", tool.FileArgs[0].Name)
			}
			if tool.FileArgs[0].Required != tt.requiredFile {
				t.Fatalf("FileArg required = %v, want %v", tool.FileArgs[0].Required, tt.requiredFile)
			}

			argNames := make(map[string]Arg)
			for _, arg := range tool.Args {
				argNames[arg.Name] = arg
			}
			if _, ok := argNames["body"]; ok {
				t.Fatal("FileRef tool should not keep generic body arg")
			}
			fileArg, ok := argNames["file"]
			if !ok {
				t.Fatal("expected file URI arg")
			}
			if fileArg.Required != tt.requiredFile {
				t.Fatalf("file arg required = %v, want %v", fileArg.Required, tt.requiredFile)
			}
			if fileArg.Schema == nil || fileArg.Schema.Format != "uri" {
				t.Fatalf("file arg schema = %+v, want uri format", fileArg.Schema)
			}
			if _, ok := argNames["note"]; ok != tt.wantNote {
				t.Fatalf("note form arg present = %v, want %v", ok, tt.wantNote)
			}
		})
	}
}

const testSpecJSONBinaryBody = `openapi: "3.0.3"
info:
  title: JSON Binary Body API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /json-only:
    post:
      operationId: uploadReportJSONOnly
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                file:
                  anyOf:
                    - type: string
                      format: binary
                    - type: string
                note:
                  type: string
      responses:
        "200":
          description: OK
`

func TestJSONBinaryBodyDoesNotInferMultipart(t *testing.T) {
	parser := NewParser(false)
	if err := parser.Parse([]byte(testSpecJSONBinaryBody)); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if len(config.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(config.Tools))
	}
	tool := config.Tools[0]
	if len(tool.FileArgs) != 0 {
		t.Fatalf("JSON request bodies should not infer multipart FileArgs, got %+v", tool.FileArgs)
	}
	if tool.UploadContentType != "" {
		t.Fatalf("JSON request bodies should not use upload content type, got %q", tool.UploadContentType)
	}
	foundBody := false
	for _, arg := range tool.Args {
		if arg.Name == "body" {
			foundBody = true
			break
		}
	}
	if !foundBody {
		t.Fatal("JSON request body should remain a regular body arg")
	}
}

// testSpecSimpleBinaryMultipart is a multipart spec with a plain binary string
// schema (not an object). With the unified file-ref approach, this now uses
// FileArgs (single "file" arg) even for plain binary multipart.
const testSpecSimpleBinaryMultipart = `openapi: "3.0.3"
info:
  title: Simple Binary Upload
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /raw-upload:
    post:
      operationId: rawUpload
      summary: Upload a raw binary file
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: string
              format: binary
      responses:
        "200":
          description: OK
`

func TestSimpleBinaryMultipart_UsesUploadContentType(t *testing.T) {
	parser := NewParser(false)
	if err := parser.Parse([]byte(testSpecSimpleBinaryMultipart)); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if len(config.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(config.Tools))
	}
	tool := config.Tools[0]

	// Plain binary multipart now uses unified file-ref approach (FileArgs)
	if tool.UploadContentType != "" {
		t.Error("plain binary multipart should NOT have UploadContentType (uses FileArgs)")
	}

	// Should have single FileArgs entry for the "file" ref
	if len(tool.FileArgs) != 1 {
		t.Fatalf("expected 1 FileArg for plain binary multipart, got %d", len(tool.FileArgs))
	}
	if tool.FileArgs[0].Name != "file" {
		t.Errorf("expected FileArgs[0].Name = 'file', got %q", tool.FileArgs[0].Name)
	}

	// Should have "file" arg (URI type), NOT file_name/file_content
	hasFile := false
	for _, arg := range tool.Args {
		if arg.Name == "file" {
			hasFile = true
			if !arg.Required {
				t.Error("file arg should be required")
			}
		}
		if arg.Name == "file_name" {
			t.Error("should NOT have legacy file_name arg")
		}
		if arg.Name == "file_content" {
			t.Error("should NOT have legacy file_content arg")
		}
	}
	if !hasFile {
		t.Error("plain binary multipart should have 'file' URI arg")
	}
}

// testSpecDuplicateOpIDs is a spec where two paths share the same operationId.
const testSpecDuplicateOpIDs = `openapi: "3.0.3"
info:
  title: API with Duplicate OperationIds
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
paths:
  /agile/sprint/{sprintId}/properties/{propertyKey}:
    delete:
      operationId: deleteProperty_1
      summary: Delete sprint property
      responses:
        "204":
          description: OK
  /dashboard/{dashboardId}/items/{itemId}/properties/{propertyKey}:
    delete:
      operationId: deleteProperty_1
      summary: Delete dashboard item property
      responses:
        "204":
          description: OK
  /agile/issue/{issueIdOrKey}:
    get:
      operationId: getIssue
      summary: Get issue (agile)
      responses:
        "200":
          description: OK
  /api/issue/{issueIdOrKey}:
    get:
      operationId: getIssue
      summary: Get issue (api)
      responses:
        "200":
          description: OK
`

func TestConverter_DuplicateOperationIds_GetUniqueToolNames(t *testing.T) {
	parser := NewParser(false)
	if err := parser.Parse([]byte(testSpecDuplicateOpIDs)); err != nil {
		t.Fatalf("failed to parse OpenAPI: %v", err)
	}

	c, err := NewConverter(parser, nil, nil, false)
	if err != nil {
		t.Fatalf("NewConverter failed: %v", err)
	}
	config, err := c.Convert()
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Should have 4 tools (2 pairs of duplicate operationIds)
	if len(config.Tools) != 4 {
		t.Fatalf("expected 4 tools, got %d", len(config.Tools))
	}

	// All tool names must be unique
	seen := make(map[string]bool)
	for _, tool := range config.Tools {
		if seen[tool.Name] {
			t.Errorf("duplicate tool name: %s", tool.Name)
		}
		seen[tool.Name] = true
	}
}
