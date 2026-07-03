package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the ContentsWithType1 tool
const ContentsWithType1InputSchema = "{\n  \"properties\": {\n    \"depth\": {\n      \"description\": \"a string indicating if all content, or just the root content of the space is returned. Default value: \\u003ccode\\u003eall\\u003c/code\\u003e. Valid values: \\u003ccode\\u003eall, root\\u003c/code\\u003e.\",\n      \"type\": \"string\"\n    },\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the space.\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 25,\n      \"description\": \"the limit of the number of labels to return, this may be restricted by fixed system limits.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"spaceKey\": {\n      \"description\": \"the key of the space to update.\",\n      \"type\": \"string\"\n    },\n    \"start\": {\n      \"description\": \"he start point of the collection to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"type\": {\n      \"description\": \"the type of content to return with the space. Valid values: \\u003ccode\\u003epage, blogpost\\u003c/code\\u003e.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\",\n    \"type\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the ContentsWithType1 tool (Status: 200, Content-Type: application/json)
const ContentsWithType1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of a piece of content.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **_links** (Type: object):\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n"

// Response Template for the ContentsWithType1 tool (Status: 404, Content-Type: application/json)
const ContentsWithType1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no content with the given id, or if the calling user does not have permission to view the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewContentsWithType1MCPTool creates the MCP Tool instance for ContentsWithType1
func NewContentsWithType1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ContentsWithType1",
		"Get contents by type - Returns the content in this given space with the given type. \n\nExample request URI: \n\n"+"\x60"+"http://example.com/confluence/rest/api/space/TEST/content/page?expand=history"+"\x60"+"",
		[]byte(ContentsWithType1InputSchema),
	)
}

// ContentsWithType1Handler is the handler function for the ContentsWithType1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ContentsWithType1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/space/{spaceKey}/content/{type}", args, []string{"spaceKey", "type"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ContentsWithType1")
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
