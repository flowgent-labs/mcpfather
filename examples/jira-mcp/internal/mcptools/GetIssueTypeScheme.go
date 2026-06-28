package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetIssueTypeScheme tool
const GetIssueTypeSchemeInputSchema = "{\n  \"properties\": {\n    \"schemeId\": {\n      \"description\": \"A String containing an issue type scheme's id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"schemeId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetIssueTypeScheme tool (Status: 200, Content-Type: application/json)
const GetIssueTypeSchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full representation of the issue type scheme with the given id.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **expand** (Type: string):\n      - Example: 'issueTypes'\n  - **id** (Type: string):\n      - Example: '10100'\n  - **issueTypes** (Type: array):\n    - **Items** (Type: object):\n      - **id** (Type: string):\n          - Example: '1'\n      - **name** (Type: string):\n          - Example: 'Bug'\n      - **self** (Type: string):\n          - Example: 'http://www.example.com/jira/rest/api/2/issuetype/1'\n      - **subtask** (Type: boolean):\n          - Example: 'false'\n      - **avatarId** (Type: integer, int64):\n          - Example: '10002'\n      - **description** (Type: string):\n          - Example: 'A problem which impairs or prevents the functions of the product.'\n      - **iconUrl** (Type: string):\n          - Example: 'http://www.example.com/jira/images/icons/issuetypes/bug.png'\n  - **name** (Type: string):\n      - Example: 'Some grouping of issue types for the greater good.'\n  - **self** (Type: string):\n      - Example: 'http://localhost:8090/jira/rest/api/2/issuetypescheme/10100'\n  - **[cyclic reference]**\n  - **description** (Type: string):\n      - Example: 'A brief explanation of this issue type scheme.'\n"

// NewGetIssueTypeSchemeMCPTool creates the MCP Tool instance for GetIssueTypeScheme
func NewGetIssueTypeSchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssueTypeScheme",
		"Get full representation of issue type scheme by id - Returns a full representation of the issue type scheme that has the given id",
		[]byte(GetIssueTypeSchemeInputSchema),
	)
}

// GetIssueTypeSchemeHandler is the handler function for the GetIssueTypeScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssueTypeSchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issuetypescheme/{schemeId}", args, []string{"schemeId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetIssueTypeScheme")
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
