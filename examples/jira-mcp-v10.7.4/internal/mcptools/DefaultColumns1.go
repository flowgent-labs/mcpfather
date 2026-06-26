package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the DefaultColumns1 tool
const DefaultColumns1InputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"The filter id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the DefaultColumns1 tool (Status: 200, Content-Type: application/json)
const DefaultColumns1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of columns for configured for the given user\n\n## Response Structure\n\n- Structure (Type: object):\n  - **columnConfig** (Type: string):\n      - Enum: ['SYSTEM', 'EXPLICIT', 'FILTER', 'USER', 'NONE']\n  - **columnLayoutItems** (Type: array):\n    - **Items** (Type: object):\n      - **navigableField** (Type: object):\n        - **columnCssClass** (Type: string):\n        - **columnHeadingKey** (Type: string):\n        - **nameKey** (Type: string):\n        - **valueLoader** (Type: object):\n          - **comparator** (Type: object):\n        - **hiddenFieldId** (Type: string):\n        - **name** (Type: string):\n        - **sortComparatorSource** (Type: object):\n        - **defaultSortOrder** (Type: string):\n        - **id** (Type: string):\n        - **sorter** (Type: object):\n          - **comparator** (Type: object):\n          - **documentConstant** (Type: string):\n      - **position** (Type: integer, int32):\n      - **columnHeadingKey** (Type: string):\n      - **id** (Type: string):\n"

// NewDefaultColumns1MCPTool creates the MCP Tool instance for DefaultColumns1
func NewDefaultColumns1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"DefaultColumns1",
		"Get default columns for filter - Returns the default columns for the given filter. Currently logged in user will be used as the user making such request.",
		[]byte(DefaultColumns1InputSchema),
	)
}

// DefaultColumns1Handler is the handler function for the DefaultColumns1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func DefaultColumns1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/filter/{id}/columns", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "DefaultColumns1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
