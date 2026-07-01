package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the DescendantsOfType tool
const DescendantsOfTypeInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the descendants.\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 25,\n      \"description\": \" (optional, default: site limit) how many items should be returned after the start index.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"description\": \"(optional, default: 0) the index of the first item within the result set that should be returned.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"type\": {\n      \"description\": \" content type to filter descendants on.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"type\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the DescendantsOfType tool (Status: 200, Content-Type: application/json)
const DescendantsOfTypeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON map representing multiple ordered collections of content descendants, keyed by content type.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n"

// Response Template for the DescendantsOfType tool (Status: 404, Content-Type: application/json)
const DescendantsOfTypeResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no content with the given id, or if the calling user does not have permission to view the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewDescendantsOfTypeMCPTool creates the MCP Tool instance for DescendantsOfType
func NewDescendantsOfTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DescendantsOfType",
		"Get descendants of type - Returns the direct descendants of a piece of Content. The ContentType(s) of the descendants returned is specified by the "+"\x60"+"type"+"\x60"+" path parameter in the request. Currently the only supported descendants are comment descendants of non-comment Content. \n\nExample request URI(s): \n\n"+"\x60"+"http://example.com/confluence/rest/api/content/1234/descendant/comment"+"\x60"+" \n\n"+"\x60"+"http://example.com/confluence/rest/api/content/1234/descendant/comment?expand=body.VIEW"+"\x60"+" \n\n"+"\x60"+"http://example.com/confluence/rest/api/content/1234/descendant/comment?start=20&limit=10"+"\x60"+"",
		[]byte(DescendantsOfTypeInputSchema),
	)
}

// DescendantsOfTypeHandler is the handler function for the DescendantsOfType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DescendantsOfTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/{id}/descendant/{type}", args, []string{"id", "type"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "DescendantsOfType")
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
