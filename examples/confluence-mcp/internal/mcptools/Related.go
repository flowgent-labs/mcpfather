package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Related tool
const RelatedInputSchema = "{\n  \"properties\": {\n    \"labelName\": {\n      \"description\": \"the name of the label (namespace prefixes permitted).\",\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 100,\n      \"description\": \"the maximum number of related labels to return. Default to be 100.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"description\": \"the start point of the collection to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"labelName\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Related tool (Status: 200, Content-Type: application/json)
const RelatedResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Return a JSON representation of related labels to the given label name\n\n## Response Structure\n\n- Structure (Type: object):\n  - **_links** (Type: object):\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n"

// Response Template for the Related tool (Status: 400, Content-Type: application/json)
const RelatedResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Return a bad request error if the given label name is invalid\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Related tool (Status: 404, Content-Type: application/json)
const RelatedResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Return a not found error if the given label name is not found\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewRelatedMCPTool creates the MCP Tool instance for Related
func NewRelatedMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Related",
		"Get related labels. - Return a paginated list of labels related to the given label name sorted by frequency of use in descending order.\nThe current process for identifying related labels solely\nexamines global labels, but it may change in the future.\n\nThe max number of labels that the API can respond with is limited, as we are filtering the access for each label.\nThis is set to 10000 by default but can be modified by the system property "+"\x60"+"confluence.rest.labels.related.max.to.process"+"\x60"+".\n\nExample request URI's:\n"+"\x60"+"https://example.com/confluence/rest/api/label/test_label_name/related"+"\x60"+"\n"+"\x60"+"https://example.com/confluence/rest/api/label/my:test_label_name/related?limit=200"+"\x60"+"",
		[]byte(RelatedInputSchema),
	)
}

// RelatedHandler is the handler function for the Related tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RelatedHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/label/{labelName}/related", args, []string{"labelName"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "Related")
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
