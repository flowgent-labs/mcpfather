package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetIssueSecurityScheme tool
const GetIssueSecuritySchemeInputSchema = "{\n  \"properties\": {\n    \"id\": {\n      \"description\": \"The issue security scheme id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetIssueSecurityScheme tool (Status: 200, Content-Type: application/json)
const GetIssueSecuritySchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns the issue security scheme with the given id.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **name** (Type: string):\n      - Example: 'My Security Scheme'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/issuesecurityschemes/10000'\n  - **defaultSecurityLevelId** (Type: integer, int64):\n      - Example: '10001'\n  - **description** (Type: string):\n      - Example: 'This is a security scheme'\n  - **id** (Type: integer, int64):\n      - Example: '10000'\n  - **levels** (Type: array):\n    - **Items** (Type: object):\n      - **name** (Type: string):\n          - Example: 'My Security Level'\n      - **self** (Type: string):\n          - Example: 'http://www.example.com/jira/rest/api/2/securitylevel/10000'\n      - **description** (Type: string):\n          - Example: 'This is a security level'\n      - **id** (Type: string):\n          - Example: '10000'\n"

// NewGetIssueSecuritySchemeMCPTool creates the MCP Tool instance for GetIssueSecurityScheme
func NewGetIssueSecuritySchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssueSecurityScheme",
		"Get specific issue security scheme by id - Returns the issue security scheme along with that are defined.",
		[]byte(GetIssueSecuritySchemeInputSchema),
	)
}

// GetIssueSecuritySchemeHandler is the handler function for the GetIssueSecurityScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssueSecuritySchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issuesecurityschemes/{id}", args, []string{"id"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetIssueSecurityScheme"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
