package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetUsersFromGroup tool
const GetUsersFromGroupInputSchema = "{\n  \"properties\": {\n    \"groupname\": {\n      \"description\": \"The group name.\",\n      \"type\": \"string\"\n    },\n    \"includeInactiveUsers\": {\n      \"description\": \"Include inactive users.\",\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"description\": \"The maximum number of users to return.\",\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"description\": \"The index of the first user in group to return.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"groupname\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetUsersFromGroup tool (Status: 200, Content-Type: application/json)
const GetUsersFromGroupResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a paginated list of users in the group\n\n## Response Structure\n\n- Structure (Type: object):\n  - **emailAddress** (Type: string):\n      - Example: 'fred@example.com'\n  - **key** (Type: string):\n      - Example: 'fred'\n  - **name** (Type: string):\n      - Example: 'Fred'\n  - **self** (Type: string):\n      - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n  - **timeZone** (Type: string):\n      - Example: 'Australia/Sydney'\n  - **active** (Type: boolean):\n      - Example: 'true'\n  - **avatarUrls** (Type: object):\n      - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n    - **Additional Properties**:\n      - **property value** (Type: string):\n          - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n  - **displayName** (Type: string):\n      - Example: 'Fred F. User'\n"

// NewGetUsersFromGroupMCPTool creates the MCP Tool instance for GetUsersFromGroup
func NewGetUsersFromGroupMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetUsersFromGroup",
		"Get users from a specified group - Returns a paginated list of users who are members of the specified group and its subgroups",
		[]byte(GetUsersFromGroupInputSchema),
	)
}

// GetUsersFromGroupHandler is the handler function for the GetUsersFromGroup tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetUsersFromGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/group/member", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetUsersFromGroup"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
