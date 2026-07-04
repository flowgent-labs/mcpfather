package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the FindGroups tool
const FindGroupsInputSchema = "{\n  \"properties\": {\n    \"exclude\": {\n      \"description\": \"List of groups to exclude\",\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"description\": \"Maximum number of results to return\",\n      \"type\": \"string\"\n    },\n    \"query\": {\n      \"description\": \"A String to match groups against\",\n      \"type\": \"string\"\n    },\n    \"userName\": {\n      \"description\": \"Username for the context\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the FindGroups tool (Status: 200, Content-Type: application/json)
const FindGroupsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a collection of matching groups\n\n## Response Structure\n\n- Structure (Type: object):\n  - **groups** (Type: array):\n    - **Items** (Type: object):\n      - **html** (Type: string):\n          - Example: '<b>j</b>dog-developers'\n      - **labels** (Type: array):\n        - **Items** (Type: object):\n          - **text** (Type: string):\n              - Example: 'jdog-developers'\n          - **title** (Type: string):\n              - Example: 'Developers'\n          - **type** (Type: string):\n              - Example: 'SINGLE'\n              - Enum: ['ADMIN', 'SINGLE', 'MULTIPLE']\n      - **name** (Type: string):\n          - Example: 'jdog-developers'\n  - **header** (Type: string):\n      - Example: 'Showing 20 of 25 matching groups'\n  - **total** (Type: integer, int32):\n      - Example: '25'\n"

// NewFindGroupsMCPTool creates the MCP Tool instance for FindGroups
func NewFindGroupsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"FindGroups",
		"Get groups matching a query - Returns groups with substrings matching a given query",
		[]byte(FindGroupsInputSchema),
	)
}

// FindGroupsHandler is the handler function for the FindGroups tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func FindGroupsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/groups/picker", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "FindGroups")
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
