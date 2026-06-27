package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Index4 tool
const Index4InputSchema = "{\n  \"properties\": {\n    \"limit\": {\n      \"default\": 25,\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"spaceKey\": {\n      \"type\": \"string\"\n    },\n    \"start\": {\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the Index4 tool (Status: 200, Content-Type: application/json)
const Index4ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns a paginated list of users watching the given Space identified by spaceKey\n\n## Response Structure\n\n- Structure (Type: object):\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n"

// Response Template for the Index4 tool (Status: 401, Content-Type: application/json)
const Index4ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> The user is not authenticated or does not have the <b>LICENSED</b> permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Index4 tool (Status: 403, Content-Type: application/json)
const Index4ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> The user is not a Confluence Administrator or Space Administrator.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the Index4 tool (Status: 404, Content-Type: application/json)
const Index4ResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> The Space does not exist.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewIndex4MCPTool creates the MCP Tool instance for Index4
func NewIndex4MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Index4",
		"Fetch users watching space - Returns a paginated list of users watching the given Space identified by spaceKey. Only a Confluence Administrator or Space Administrator can perform this action.",
		[]byte(Index4InputSchema),
	)
}

// Index4Handler is the handler function for the Index4 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func Index4Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/space/{spaceKey}/watchers", args, []string{"spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Index4"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
