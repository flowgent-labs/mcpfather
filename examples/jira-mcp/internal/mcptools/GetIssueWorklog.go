package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetIssueWorklog tool
const GetIssueWorklogInputSchema = "{\n  \"properties\": {\n    \"issueIdOrKey\": {\n      \"description\": \"Issue id or key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"issueIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetIssueWorklog tool (Status: 200, Content-Type: application/json)
const GetIssueWorklogResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a collection of worklogs associated with the issue, with count and pagination information.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **worklogs** (Type: array):\n    - **Items** (Type: object):\n      - **started** (Type: string):\n          - Example: '2010-07-14T18:23:23.733+0000'\n      - **updateAuthor** (Type: object):\n        - **emailAddress** (Type: string):\n            - Example: 'fred@example.com'\n        - **key** (Type: string):\n            - Example: 'fred'\n        - **name** (Type: string):\n            - Example: 'Fred'\n        - **self** (Type: string):\n            - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n        - **timeZone** (Type: string):\n            - Example: 'Australia/Sydney'\n        - **active** (Type: boolean):\n            - Example: 'true'\n        - **avatarUrls** (Type: object):\n            - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n          - **Additional Properties**:\n            - **property value** (Type: string):\n                - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n        - **displayName** (Type: string):\n            - Example: 'Fred F. User'\n      - **visibility** (Type: object):\n        - **type** (Type: string):\n            - Example: 'group'\n            - Enum: ['group', 'role']\n        - **value** (Type: string):\n            - Example: 'jira-software-users'\n      - **[cyclic reference]**\n      - **comment** (Type: string):\n          - Example: 'I did some work here.'\n      - **timeSpent** (Type: string):\n          - Example: '3h 20m'\n      - **issueId** (Type: string):\n          - Example: '10002'\n      - **self** (Type: string, uri):\n          - Example: 'http://www.example.com/jira/rest/api/2/issue/10010/worklog/10000'\n      - **created** (Type: string):\n          - Example: '2010-07-14T18:23:23.733+0000'\n      - **id** (Type: string):\n          - Example: '100028'\n      - **timeSpentSeconds** (Type: integer, int64):\n          - Example: '12000'\n      - **updated** (Type: string):\n          - Example: '2010-07-14T18:23:23.733+0000'\n  - **maxResults** (Type: integer, int32):\n      - Example: '1'\n  - **startAt** (Type: integer, int32):\n      - Example: '0'\n  - **total** (Type: integer, int32):\n      - Example: '1'\n"

// NewGetIssueWorklogMCPTool creates the MCP Tool instance for GetIssueWorklog
func NewGetIssueWorklogMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIssueWorklog",
		"Get worklogs for an issue - Returns all work logs for an issue. Work logs won't be returned if the Log work field is hidden for the project.",
		[]byte(GetIssueWorklogInputSchema),
	)
}

// GetIssueWorklogHandler is the handler function for the GetIssueWorklog tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIssueWorklogHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/issue/{issueIdOrKey}/worklog", args, []string{"issueIdOrKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetIssueWorklog")
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
