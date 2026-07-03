package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Descendants tool
const DescendantsInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \" a comma separated list of properties to expand on the descendants.\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Descendants tool (Status: 200, Content-Type: application/json)
const DescendantsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON map representing multiple ordered collections of content descendants,keyed by content type.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **_links** (Type: object):\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n"

// Response Template for the Descendants tool (Status: 404, Content-Type: application/json)
const DescendantsResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no content with the given id, or if the calling user does not have permission to view the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewDescendantsMCPTool creates the MCP Tool instance for Descendants
func NewDescendantsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Descendants",
		"Get Descendants - Returns a map of the descendants of a piece of Content. Content can have multiple types of descendants - for example a Page can have descendants that are also Pages, but it can also have Comments and Attachments. \n\nThe ContentType(s) of the descendants returned is specified by the "+"\x60"+"expand"+"\x60"+" query parameter in the request - this parameter can include expands for multiple descendant types. If no types are included in the expand parameter, the map returned will just list the descendant types that are available to be expanded for the Content referenced by the "+"\x60"+"id"+"\x60"+" path parameter. \n\nCurrently the only supported descendants are comment descendants of non-comment Content. \n\nExample request URI(s): \n\n"+"\x60"+"http://example.com/confluence/rest/api/content/1234/descendant"+"\x60"+" \n\n"+"\x60"+"http://example.com/confluence/rest/api/content/1234/descendant?expand=comment.body.VIEW"+"\x60"+" \n\n"+"\x60"+"http://example.com/confluence/rest/api/content/1234/descendant?expand=comment"+"\x60"+"",
		[]byte(DescendantsInputSchema),
	)
}

// DescendantsHandler is the handler function for the Descendants tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DescendantsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/{id}/descendant", args, []string{"id"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Descendants")
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
