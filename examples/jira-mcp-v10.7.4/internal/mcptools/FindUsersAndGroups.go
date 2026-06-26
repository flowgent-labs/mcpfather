package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the FindUsersAndGroups tool
const FindUsersAndGroupsInputSchema = "{\n  \"properties\": {\n    \"fieldId\": {\n      \"description\": \"The custom field id\",\n      \"type\": \"string\"\n    },\n    \"issueTypeId\": {\n      \"description\": \"The list of issue type ids to further restrict the search\",\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"description\": \"The maximum number of users to return\",\n      \"type\": \"string\"\n    },\n    \"projectId\": {\n      \"description\": \"The list of project ids to further restrict the search\",\n      \"type\": \"string\"\n    },\n    \"query\": {\n      \"description\": \"A string used to search username, Name or e-mail address\",\n      \"type\": \"string\"\n    },\n    \"showAvatar\": {\n      \"description\": \"Show avatar\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the FindUsersAndGroups tool (Status: 200, Content-Type: application/json)
const FindUsersAndGroupsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of users and groups matching query with highlighting\n\n## Response Structure\n\n- Structure (Type: object):\n  - **groups** (Type: object):\n    - **groups** (Type: array):\n      - **Items** (Type: object):\n        - **html** (Type: string):\n            - Example: '<b>j</b>dog-developers'\n        - **labels** (Type: array):\n          - **Items** (Type: object):\n            - **text** (Type: string):\n                - Example: 'jdog-developers'\n            - **title** (Type: string):\n                - Example: 'Developers'\n            - **type** (Type: string):\n                - Example: 'SINGLE'\n                - Enum: ['ADMIN', 'SINGLE', 'MULTIPLE']\n        - **name** (Type: string):\n            - Example: 'jdog-developers'\n    - **header** (Type: string):\n        - Example: 'Showing 20 of 25 matching groups'\n    - **total** (Type: integer, int32):\n        - Example: '25'\n  - **users** (Type: object):\n"

// NewFindUsersAndGroupsMCPTool creates the MCP Tool instance for FindUsersAndGroups
func NewFindUsersAndGroupsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"FindUsersAndGroups",
		"Get users and groups matching query with highlighting - Returns a list of users and groups matching query with highlighting",
		[]byte(FindUsersAndGroupsInputSchema),
	)
}

// FindUsersAndGroupsHandler is the handler function for the FindUsersAndGroups tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func FindUsersAndGroupsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/groupuserpicker", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "FindUsersAndGroups"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
