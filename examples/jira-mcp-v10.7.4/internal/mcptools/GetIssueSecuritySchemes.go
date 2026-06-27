package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetIssueSecuritySchemes tool
const GetIssueSecuritySchemesInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetIssueSecuritySchemes tool (Status: 200, Content-Type: application/json)
const GetIssueSecuritySchemesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of all available issue security schemes.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **issueSecuritySchemes** (Type: array):\n    - **Items** (Type: object):\n      - **levels** (Type: array):\n        - **Items** (Type: object):\n          - **name** (Type: string):\n              - Example: 'My Security Level'\n          - **self** (Type: string):\n              - Example: 'http://www.example.com/jira/rest/api/2/securitylevel/10000'\n          - **description** (Type: string):\n              - Example: 'This is a security level'\n          - **id** (Type: string):\n              - Example: '10000'\n      - **name** (Type: string):\n          - Example: 'My Security Scheme'\n      - **self** (Type: string):\n          - Example: 'http://www.example.com/jira/rest/api/2/issuesecurityschemes/10000'\n      - **defaultSecurityLevelId** (Type: integer, int64):\n          - Example: '10001'\n      - **description** (Type: string):\n          - Example: 'This is a security scheme'\n      - **id** (Type: integer, int64):\n          - Example: '10000'\n"

// NewGetIssueSecuritySchemesMCPTool creates the MCP Tool instance for GetIssueSecuritySchemes
func NewGetIssueSecuritySchemesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssueSecuritySchemes",
		"Get all issue security schemes - Returns all issue security schemes that are defined.",
		[]byte(GetIssueSecuritySchemesInputSchema),
	)
}

// GetIssueSecuritySchemesHandler is the handler function for the GetIssueSecuritySchemes tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssueSecuritySchemesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issuesecurityschemes", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetIssueSecuritySchemes"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
