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
	if uploadTool.UploadContentType == "" {
		t.Error("expected uploadAttachment to have UploadContentType set")
	}

	if downloadTool == nil {
		t.Fatal("expected downloadAttachment tool to be generated")
	}
	if downloadTool.UploadContentType != "" {
		t.Error("download tool should not have UploadContentType set")
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
