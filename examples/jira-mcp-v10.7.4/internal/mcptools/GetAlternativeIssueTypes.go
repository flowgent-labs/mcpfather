package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAlternativeIssueTypes tool
const GetAlternativeIssueTypesInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"The issue type id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAlternativeIssueTypes tool (Status: 200, Content-Type: application/json)
const GetAlternativeIssueTypesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of issue types.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **description** (Type: string):\n      - Example: 'A problem which impairs or prevents the functions of the product.'\n  - **iconUrl** (Type: string):\n      - Example: 'http://www.example.com/jira/images/icons/issuetypes/bug.png'\n  - **id** (Type: string):\n      - Example: '1'\n  - **name** (Type: string):\n      - Example: 'Bug'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/issuetype/1'\n  - **subtask** (Type: boolean):\n      - Example: 'false'\n  - **avatarId** (Type: integer, int64):\n      - Example: '10002'\n"

// NewGetAlternativeIssueTypesMCPTool creates the MCP Tool instance for GetAlternativeIssueTypes
func NewGetAlternativeIssueTypesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAlternativeIssueTypes",
		"Get list of alternative issue types for given id - Returns a list of all alternative issue types for the given issue type id.",
		[]byte(GetAlternativeIssueTypesInputSchema),
	)
}

// GetAlternativeIssueTypesHandler is the handler function for the GetAlternativeIssueTypes tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAlternativeIssueTypesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issuetype/{id}/alternatives", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAlternativeIssueTypes"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
