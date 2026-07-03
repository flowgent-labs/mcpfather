package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetStatusCategory tool
const GetStatusCategoryInputSchema = "{\n  \"properties\": {\n    \"idOrKey\": {\n      \"description\": \"A numeric StatusCategory id or a status category key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"idOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetStatusCategory tool (Status: 200, Content-Type: application/json)
const GetStatusCategoryResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full representation of a Jira issue status category in JSON format.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **key** (Type: string):\n      - Example: 'new'\n  - **name** (Type: string):\n      - Example: 'To Do'\n  - **self** (Type: string):\n      - Example: 'http://localhost:8090/jira/rest/api/2.0/statuscategory/1'\n  - **colorName** (Type: string):\n      - Example: 'blue-gray'\n  - **id** (Type: integer, int64):\n      - Example: '1'\n"

// NewGetStatusCategoryMCPTool creates the MCP Tool instance for GetStatusCategory
func NewGetStatusCategoryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetStatusCategory",
		"Get status category by ID or key - Returns a full representation of the StatusCategory having the given id or key",
		[]byte(GetStatusCategoryInputSchema),
	)
}

// GetStatusCategoryHandler is the handler function for the GetStatusCategory tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetStatusCategoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/statuscategory/{idOrKey}", args, []string{"idOrKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetStatusCategory")
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
