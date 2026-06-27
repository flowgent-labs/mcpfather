package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the ChildrenOfType tool
const ChildrenOfTypeInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"a comma separated list of properties to expand on the children\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"default\": 25,\n      \"description\": \"how many items should be returned after the start index\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"parentVersion\": {\n      \"default\": 0,\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"description\": \"the index of the first item within the result set that should be returned\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"type\": {\n      \"description\": \"a  content type to filter children on.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"type\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the ChildrenOfType tool (Status: 200, Content-Type: application/json)
const ChildrenOfTypeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON map representing multiple ordered collections of content children, keyed by content type\n\n## Response Structure\n\n- Structure (Type: object):\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n"

// Response Template for the ChildrenOfType tool (Status: 404, Content-Type: application/json)
const ChildrenOfTypeResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n>  Returned if there is no content with the given id, or if the calling user does not have permission to view the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewChildrenOfTypeMCPTool creates the MCP Tool instance for ChildrenOfType
func NewChildrenOfTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ChildrenOfType",
		"Get children of content by type - Returns the direct children of a piece of Content, limited to a single child type.The types of the children returned is specified by the \"type\" path parameter in the request.",
		[]byte(ChildrenOfTypeInputSchema),
	)
}

// ChildrenOfTypeHandler is the handler function for the ChildrenOfType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ChildrenOfTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/content/{id}/child/{type}", args, []string{"id", "type"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "ChildrenOfType"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
