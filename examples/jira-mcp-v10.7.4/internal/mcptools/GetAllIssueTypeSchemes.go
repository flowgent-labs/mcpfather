package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetAllIssueTypeSchemes tool
const GetAllIssueTypeSchemesInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetAllIssueTypeSchemes tool (Status: 200, Content-Type: application/json)
const GetAllIssueTypeSchemesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of issue type schemes.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **schemes** (Type: array):\n    - **Items** (Type: object):\n      - **defaultIssueType** (Type: object):\n        - **avatarId** (Type: integer, int64):\n            - Example: '10002'\n        - **description** (Type: string):\n            - Example: 'A problem which impairs or prevents the functions of the product.'\n        - **iconUrl** (Type: string):\n            - Example: 'http://www.example.com/jira/images/icons/issuetypes/bug.png'\n        - **id** (Type: string):\n            - Example: '1'\n        - **name** (Type: string):\n            - Example: 'Bug'\n        - **self** (Type: string):\n            - Example: 'http://www.example.com/jira/rest/api/2/issuetype/1'\n        - **subtask** (Type: boolean):\n            - Example: 'false'\n      - **description** (Type: string):\n          - Example: 'A brief explanation of this issue type scheme.'\n      - **expand** (Type: string):\n          - Example: 'issueTypes'\n      - **id** (Type: string):\n          - Example: '10100'\n      - **issueTypes** (Type: array):\n        - **[cyclic reference]**\n      - **name** (Type: string):\n          - Example: 'Some grouping of issue types for the greater good.'\n      - **self** (Type: string):\n          - Example: 'http://localhost:8090/jira/rest/api/2/issuetypescheme/10100'\n"

// NewGetAllIssueTypeSchemesMCPTool creates the MCP Tool instance for GetAllIssueTypeSchemes
func NewGetAllIssueTypeSchemesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllIssueTypeSchemes",
		"Get list of all issue type schemes visible to user - Returns a list of all issue type schemes visible to the user. All issue types associated with the scheme will only be returned if an additional query parameter is provided: expand=schemes.issueTypes. Similarly, the default issue type associated with the scheme (if one exists) will only be returned if an additional query parameter is provided: expand=schemes.defaultIssueType. Note that both query parameters can be used together: expand=schemes.issueTypes,schemes.defaultIssueType.",
		[]byte(GetAllIssueTypeSchemesInputSchema),
	)
}

// GetAllIssueTypeSchemesHandler is the handler function for the GetAllIssueTypeSchemes tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllIssueTypeSchemesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issuetypescheme", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAllIssueTypeSchemes"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
