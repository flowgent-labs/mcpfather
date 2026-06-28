package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetVersionRelatedIssues tool
const GetVersionRelatedIssuesInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"ID of the version.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetVersionRelatedIssues tool (Status: 200, Content-Type: application/json)
const GetVersionRelatedIssuesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the version exists and the currently authenticated user has permission to view it.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **customFieldNames** (Type: array):\n    - **Items** (Type: object):\n      - **customFieldId** (Type: integer, int64):\n          - Example: '10000'\n      - **fieldName** (Type: string):\n          - Example: 'Field1'\n      - **issueCountWithVersionInCustomField** (Type: integer, int64):\n          - Example: '2'\n  - **issueCountWithCustomFieldsShowingVersion** (Type: integer, int64):\n      - Example: '54'\n  - **issuesAffectedCount** (Type: integer, int64):\n      - Example: '101'\n  - **issuesFixedCount** (Type: integer, int64):\n      - Example: '23'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/version/10000'\n"

// NewGetVersionRelatedIssuesMCPTool creates the MCP Tool instance for GetVersionRelatedIssues
func NewGetVersionRelatedIssuesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetVersionRelatedIssues",
		"Get version related issues count - Returns a bean containing the number of fixed in and affected issues for the given version.",
		[]byte(GetVersionRelatedIssuesInputSchema),
	)
}

// GetVersionRelatedIssuesHandler is the handler function for the GetVersionRelatedIssues tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetVersionRelatedIssuesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/version/{id}/relatedIssueCounts", args, []string{"id"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetVersionRelatedIssues")
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
