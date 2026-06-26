package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetIssueSecurityScheme1 tool
const GetIssueSecurityScheme1InputSchema = "{\n  \"properties\": {\n    \"projectKeyOrId\": {\n      \"description\": \"The project id or project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"projectKeyOrId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetIssueSecurityScheme1 tool (Status: 200, Content-Type: application/json)
const GetIssueSecurityScheme1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Issue security scheme\n\n## Response Structure\n\n- Structure (Type: object):\n  - **description** (Type: string):\n      - Example: 'This is a security scheme'\n  - **id** (Type: integer, int64):\n      - Example: '10000'\n  - **levels** (Type: array):\n    - **Items** (Type: object):\n      - **name** (Type: string):\n          - Example: 'My Security Level'\n      - **self** (Type: string):\n          - Example: 'http://www.example.com/jira/rest/api/2/securitylevel/10000'\n      - **description** (Type: string):\n          - Example: 'This is a security level'\n      - **id** (Type: string):\n          - Example: '10000'\n  - **name** (Type: string):\n      - Example: 'My Security Scheme'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/issuesecurityschemes/10000'\n  - **defaultSecurityLevelId** (Type: integer, int64):\n      - Example: '10001'\n"

// NewGetIssueSecurityScheme1MCPTool creates the MCP Tool instance for GetIssueSecurityScheme1
func NewGetIssueSecurityScheme1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssueSecurityScheme1",
		"Get issue security scheme for project - Returns the issue security scheme for project.",
		[]byte(GetIssueSecurityScheme1InputSchema),
	)
}

// GetIssueSecurityScheme1Handler is the handler function for the GetIssueSecurityScheme1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssueSecurityScheme1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/project/{projectKeyOrId}/issuesecuritylevelscheme", args, []string{"projectKeyOrId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetIssueSecurityScheme1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
