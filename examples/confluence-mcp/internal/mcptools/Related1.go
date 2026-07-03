package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Related1 tool
const Related1InputSchema = "{\n  \"properties\": {\n    \"labelName\": {\n      \"description\": \"a string containing the name of the label\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 25,\n      \"description\": \"the limit of the number of labels to return, this may be restricted by fixed system limits\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"spaceKey\": {\n      \"description\": \"a string containing the key of the space\",\n      \"type\": \"string\"\n    },\n    \"start\": {\n      \"description\": \"the start point of the collection to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"labelName\",\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Related1 tool (Status: 200, Content-Type: application/json)
const Related1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> a JSON representation of the lists label, or an empty list if no labels are found.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n"

// Response Template for the Related1 tool (Status: 403, Content-Type: application/json)
const Related1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> If the calling user does not have permission to view the given space.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Related1 tool (Status: 404, Content-Type: application/json)
const Related1ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no space with the given spaceKey.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewRelated1MCPTool creates the MCP Tool instance for Related1
func NewRelated1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Related1",
		"Get related labels - Returns a paginated list of related Labels used by Content within the given Space.A Label is defined as being related to another when it is found attached to the same Content as the Label specified in the request.\n\nExample request URI:\n"+"\x60"+"https://example.com/confluence/rest/api/space/TEST/labels/example-label/related"+"\x60"+"",
		[]byte(Related1InputSchema),
	)
}

// Related1Handler is the handler function for the Related1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Related1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/space/{spaceKey}/labels/{labelName}/related", args, []string{"labelName", "spaceKey"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Related1")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
