package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the FindBulkAssignableUsers tool
const FindBulkAssignableUsersInputSchema = "{\n  \"properties\": {\n    \"maxResults\": {\n      \"default\": 50,\n      \"description\": \"The maximum number of users to return (defaults to 50). The maximum allowed value is 100 (The combination of maxResults and startAt is limited to the first 100 results). If you specify a value that is higher than this number, your search results will be truncated. If you send a request with startAt=98 and maxResults=20, it will only return 2 users.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"projectKeys\": {\n      \"description\": \"the keys of the projects we are finding assignable users for, comma-separated\",\n      \"type\": \"string\"\n    },\n    \"username\": {\n      \"description\": \"the username\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the FindBulkAssignableUsers tool (Status: 200, Content-Type: application/json)
const FindBulkAssignableUsersResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of users that match the search string.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **locale** (Type: string):\n      - Example: 'en_AU'\n  - **key** (Type: string):\n      - Example: 'JIRAUSER10100'\n  - **name** (Type: string):\n      - Example: 'fred'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2.0/user?username=fred'\n  - **timeZone** (Type: string):\n      - Example: 'Australia/Sydney'\n  - **lastLoginTime** (Type: string):\n      - Example: '2023-08-30T16:37:01+1000'\n  - **deleted** (Type: boolean):\n      - Example: 'false'\n  - **groups** (Type: object):\n    - **maxResults** (Type: integer, int32):\n    - **pagingCallback** (Type: object):\n    - **size** (Type: integer, int32):\n    - **[cyclic reference]**\n  - **active** (Type: boolean):\n      - Example: 'true'\n  - **applicationRoles** (Type: object):\n      - Example: '[\"jira-core\",\"jira-admin\",\"important\"]'\n    - **maxResults** (Type: integer, int32):\n    - **pagingCallback** (Type: object):\n    - **size** (Type: integer, int32):\n    - **[cyclic reference]**\n  - **avatarUrls** (Type: object):\n      - Example: '{\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\",\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\"}'\n    - **Additional Properties**:\n      - **property value** (Type: string, uri):\n          - Example: '{\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large&ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small&ownerId=fred\",\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall&ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium&ownerId=fred\"}'\n  - **displayName** (Type: string):\n      - Example: 'Fred F. User'\n  - **emailAddress** (Type: string):\n      - Example: 'fred@example.com'\n  - **expand** (Type: string):\n"

// NewFindBulkAssignableUsersMCPTool creates the MCP Tool instance for FindBulkAssignableUsers
func NewFindBulkAssignableUsersMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"FindBulkAssignableUsers",
		"Find bulk assignable users - Returns a list of users that match the search string and can be assigned issues for all the given projects.",
		[]byte(FindBulkAssignableUsersInputSchema),
	)
}

// FindBulkAssignableUsersHandler is the handler function for the FindBulkAssignableUsers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func FindBulkAssignableUsersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/assignable/multiProjectSearch", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "FindBulkAssignableUsers")
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
