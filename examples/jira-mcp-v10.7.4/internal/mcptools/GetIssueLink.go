package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetIssueLink tool
const GetIssueLinkInputSchema = "{\n  \"properties\": {\n    \"linkId\": {\n      \"description\": \"The issue link id.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"linkId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetIssueLink tool (Status: 200, Content-Type: application/json)
const GetIssueLinkResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned if the request was successful.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **outwardIssue** (Type: object):\n    - **fields** (Type: object):\n      - **issuetype** (Type: object):\n        - **name** (Type: string):\n            - Example: 'Bug'\n        - **self** (Type: string):\n            - Example: 'http://www.example.com/jira/rest/api/2/issuetype/1'\n        - **subtask** (Type: boolean):\n            - Example: 'false'\n        - **avatarId** (Type: integer, int64):\n            - Example: '10002'\n        - **description** (Type: string):\n            - Example: 'A problem which impairs or prevents the functions of the product.'\n        - **iconUrl** (Type: string):\n            - Example: 'http://www.example.com/jira/images/icons/issuetypes/bug.png'\n        - **id** (Type: string):\n            - Example: '1'\n      - **priority** (Type: object):\n        - **description** (Type: string):\n            - Example: 'This is a description of the priority'\n        - **iconUrl** (Type: string):\n            - Example: 'http://www.example.com/jira/images/icons/priorities/major.png'\n        - **id** (Type: string):\n            - Example: '1'\n        - **name** (Type: string):\n            - Example: 'Major'\n        - **self** (Type: string):\n            - Example: 'http://www.example.com/jira/rest/api/2/priority/1'\n        - **statusColor** (Type: string):\n            - Example: 'red'\n      - **status** (Type: object):\n        - **description** (Type: string):\n            - Example: 'The issue is currently being worked on.'\n        - **iconUrl** (Type: string):\n            - Example: 'http://localhost:8090/jira/images/icons/progress.gif'\n        - **id** (Type: string):\n            - Example: '10000'\n        - **name** (Type: string):\n            - Example: 'In Progress'\n        - **self** (Type: string):\n            - Example: 'http://localhost:8090/jira/rest/api/2.0/status/10000'\n        - **statusCategory** (Type: object):\n          - **self** (Type: string):\n              - Example: 'http://localhost:8090/jira/rest/api/2.0/statuscategory/1'\n          - **colorName** (Type: string):\n              - Example: 'blue-gray'\n          - **id** (Type: integer, int64):\n              - Example: '1'\n          - **key** (Type: string):\n              - Example: 'new'\n          - **name** (Type: string):\n              - Example: 'To Do'\n        - **statusColor** (Type: string):\n            - Example: 'green'\n      - **summary** (Type: string):\n    - **id** (Type: string):\n        - Example: '10000'\n    - **key** (Type: string):\n        - Example: 'HSP-1'\n    - **self** (Type: string, uri):\n        - Example: 'http://www.example.com/jira/rest/api/2/issue/10000'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/issueLink/10000'\n  - **type** (Type: object):\n    - **name** (Type: string):\n        - Example: 'Duplicate'\n    - **outward** (Type: string):\n        - Example: 'duplicates'\n    - **self** (Type: string, uri):\n        - Example: 'http://www.example.com/jira/rest/api/2/issueLinkType/10000'\n    - **id** (Type: string):\n        - Example: '10000'\n    - **inward** (Type: string):\n        - Example: 'is duplicated by'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **[cyclic reference]**\n"

// NewGetIssueLinkMCPTool creates the MCP Tool instance for GetIssueLink
func NewGetIssueLinkMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssueLink",
		"Get an issue link with the specified id - Returns an issue link with the specified id.",
		[]byte(GetIssueLinkInputSchema),
	)
}

// GetIssueLinkHandler is the handler function for the GetIssueLink tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssueLinkHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issueLink/{linkId}", args, []string{"linkId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetIssueLink"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
