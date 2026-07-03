package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Get1 tool
const Get1InputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the space properties. Default value: \\u003ccode\\u003eversion\\u003c/code\\u003e.\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 10,\n      \"description\": \"the limit of the number of items to return, this may be restricted by fixed system limits.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"spaceKey\": {\n      \"description\": \"the space to find properties under. Required.\",\n      \"type\": \"string\"\n    },\n    \"start\": {\n      \"description\": \"he start point of the collection to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Get1 tool (Status: 200, Content-Type: application/json)
const Get1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> a JSON representation of the space properties.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **_links** (Type: object):\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n"

// Response Template for the Get1 tool (Status: 404, Content-Type: application/json)
const Get1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no space with the given key, or if the calling user does not have permission to view the space.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGet1MCPTool creates the MCP Tool instance for Get1
func NewGet1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Get1",
		"Get space properties - Returns a paginated list of space properties. \n\nExample request URI: \n\n"+"\x60"+"http://example.com/confluence/rest/api/space/TST/property?expand=space,version"+"\x60"+"",
		[]byte(Get1InputSchema),
	)
}

// Get1Handler is the handler function for the Get1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Get1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/space/{spaceKey}/property", args, []string{"spaceKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Get1")
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
