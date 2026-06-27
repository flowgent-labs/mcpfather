package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the Recent tool
const RecentInputSchema = "{\n  \"properties\": {\n    \"limit\": {\n      \"default\": 100,\n      \"description\": \"the limit of the number of labels to return, this may be restricted by fixed system limits\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"start\": {\n      \"description\": \"the start point of the collection to return.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the Recent tool (Status: 200, Content-Type: application/json)
const RecentResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> a JSON representation of a list of labels, or an empty list if no labels are found.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n  - **_links** (Type: object):\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n"

// Response Template for the Recent tool (Status: 403, Content-Type: application/json)
const RecentResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> If the calling user does not have permission to retrieve recently used labels.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewRecentMCPTool creates the MCP Tool instance for Recent
func NewRecentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"Recent",
		"Get recently used labels - Returns a paginated list of recently used labels within a Confluence instance. Labels are sorted\nfrom the most to the least recently used. Only global labels are considered in this list.\n\nExample request URI's:\n"+"\x60"+"https://example.com/confluence/rest/api/label/recent"+"\x60"+"\n"+"\x60"+"https://example.com/confluence/rest/api/label/recent?start=2&limit=1"+"\x60"+"",
		[]byte(RecentInputSchema),
	)
}

// RecentHandler is the handler function for the Recent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RecentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/label/recent", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "Recent"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
