package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the ContentsWithType tool
const ContentsWithTypeInputSchema = "{\n  \"properties\": {\n    \"cursor\": {\n      \"description\": \"the identifier which is used to skip results from a previous query when paginating. Cursor is empty in first request, to move forward or backward use cursor provided in response.\",\n      \"type\": \"string\"\n    },\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the space.\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 25,\n      \"description\": \"the limit of the number of labels to return, this may be restricted by fixed system limits.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"spaceKey\": {\n      \"description\": \"the key of the space to update.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the ContentsWithType tool (Status: 200, Content-Type: application/json)
const ContentsWithTypeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns an array of full JSON representations of trash contents.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **nextCursor** (Type: string):\n      - Example: 'cursortype:false:360456'\n  - **prevCursor** (Type: string):\n      - Example: 'cursortype:true:360423'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/confluence/rest/api/latest/..?limit=25&cursor=cursortype%3Atrue%3A360423'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/confluence/rest/api/..?limit=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/confluence/rest/api/latest/..?limit=25&cursor=cursortype%3Afalse%3A360456'\n  - **cursor** (Type: string):\n      - Example: 'cursortype:false:360422'\n  - **limit** (Type: number):\n      - Example: '25'\n"

// Response Template for the ContentsWithType tool (Status: 403, Content-Type: application/json)
const ContentsWithTypeResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling user does not have admin permission of the Space.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the ContentsWithType tool (Status: 404, Content-Type: application/json)
const ContentsWithTypeResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if space not found by space key.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewContentsWithTypeMCPTool creates the MCP Tool instance for ContentsWithType
func NewContentsWithTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ContentsWithType",
		"Get trash contents of space - Returns the trash contents in this given space. \n\nExample request URI: \n\n"+"\x60"+"http://example.com/confluence/rest/api/space/TEST/trash?limit=100&cursor=content:false:612345"+"\x60"+"",
		[]byte(ContentsWithTypeInputSchema),
	)
}

// ContentsWithTypeHandler is the handler function for the ContentsWithType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ContentsWithTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/space/{spaceKey}/trash", args, []string{"spaceKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ContentsWithType")
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
