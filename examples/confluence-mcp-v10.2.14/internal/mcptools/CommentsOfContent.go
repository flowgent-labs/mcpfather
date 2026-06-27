package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the CommentsOfContent tool
const CommentsOfContentInputSchema = "{\n  \"properties\": {\n    \"depth\": {\n      \"description\": \"(optional, default: \\\"\\\") the depth of the comments. Possible values are: \\\"\\\" (ROOT only), \\\"all\\\"\",\n      \"type\": \"string\"\n    },\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the children\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 25,\n      \"description\": \"how many items should be returned after the start index\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"location\": {\n      \"description\": \"(optional, default: \\\"\\\" means all) the location of the comments. Possible values are: \\\"inline\\\", \\\"footer\\\", \\\"resolved\\\".\\nYou can define multiple location params. The results will be the comments matched by any location.\",\n      \"items\": {\n        \"type\": \"string\"\n      },\n      \"type\": \"array\",\n      \"uniqueItems\": true\n    },\n    \"parentVersion\": {\n      \"default\": 0,\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"description\": \"the index of the first item within the result set that should be returned\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the CommentsOfContent tool (Status: 200, Content-Type: application/json)
const CommentsOfContentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON map representing multiple ordered collections of content children, keyed by content type\n\n## Response Structure\n\n- Structure (Type: object):\n  - **_links** (Type: object):\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n"

// Response Template for the CommentsOfContent tool (Status: 404, Content-Type: application/json)
const CommentsOfContentResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no content with the given id, or if the calling user does not have permission to\nview the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewCommentsOfContentMCPTool creates the MCP Tool instance for CommentsOfContent
func NewCommentsOfContentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CommentsOfContent",
		"Get comments of content - Returns the comments of a piece of Content. Example request URI(s): \n\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/child/comment"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/child/comment?expand=body.view"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/child/comment?start=20&limit=10"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/child/comment?location=footer&location=inline&location=resolved"+"\x60"+"\n- "+"\x60"+"http://example.com/confluence/rest/api/content/1234/child/comment?expand=extensions.inlineProperties,extensions.resolution"+"\x60"+"",
		[]byte(CommentsOfContentInputSchema),
	)
}

// CommentsOfContentHandler is the handler function for the CommentsOfContent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CommentsOfContentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/{id}/child/comment", args, []string{"id"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CommentsOfContent"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
