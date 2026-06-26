package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetIssueType tool
const GetIssueTypeInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"The id of the scheme.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"issueType\": {\n      \"description\": \"The issue type to query.\",\n      \"type\": \"string\"\n    },\n    \"returnDraftIfExists\": {\n      \"default\": false,\n      \"description\": \"When true indicates that a scheme's draft, if it exists, should be queried instead of the scheme itself.\",\n      \"type\": \"boolean\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"issueType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetIssueType tool (Status: 200, Content-Type: application/json)
const GetIssueTypeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned on success.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **issueType** (Type: string):\n      - Example: '10000'\n  - **updateDraftIfNeeded** (Type: boolean):\n      - Example: 'false'\n  - **workflow** (Type: string):\n      - Example: 'My Workflow'\n"

// NewGetIssueTypeMCPTool creates the MCP Tool instance for GetIssueType
func NewGetIssueTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssueType",
		"Get issue type mapping for a scheme - Returns the issue type mapping for the passed workflow scheme.",
		[]byte(GetIssueTypeInputSchema),
	)
}

// GetIssueTypeHandler is the handler function for the GetIssueType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssueTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/workflowscheme/{id}/issuetype/{issueType}", args, []string{"id", "issueType"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetIssueType"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
