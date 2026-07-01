package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetDraftIssueType tool
const GetDraftIssueTypeInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"The id of the parent scheme.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"issueType\": {\n      \"description\": \"The issue type to query.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\",\n    \"issueType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetDraftIssueType tool (Status: 200, Content-Type: application/json)
const GetDraftIssueTypeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned on success.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **issueType** (Type: string):\n      - Example: '10000'\n  - **updateDraftIfNeeded** (Type: boolean):\n      - Example: 'false'\n  - **workflow** (Type: string):\n      - Example: 'My Workflow'\n"

// NewGetDraftIssueTypeMCPTool creates the MCP Tool instance for GetDraftIssueType
func NewGetDraftIssueTypeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetDraftIssueType",
		"Get issue type mapping for a draft scheme - Returns the issue type mapping for the passed draft workflow scheme.",
		[]byte(GetDraftIssueTypeInputSchema),
	)
}

// GetDraftIssueTypeHandler is the handler function for the GetDraftIssueType tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetDraftIssueTypeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/workflowscheme/{id}/draft/issuetype/{issueType}", args, []string{"id", "issueType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetDraftIssueType")
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
