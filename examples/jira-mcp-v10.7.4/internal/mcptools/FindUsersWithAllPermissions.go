package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the FindUsersWithAllPermissions tool
const FindUsersWithAllPermissionsInputSchema = "{\n  \"properties\": {\n    \"issueKey\": {\n      \"description\": \"the issue key for the issue for which returned users have specified permissions.\",\n      \"type\": \"string\"\n    },\n    \"maxResults\": {\n      \"description\": \"The maximum number of users to return (defaults to 50). The maximum allowed value is 100 (The combination of maxResults and startAt is limited to the first 100 results). If you specify a value that is higher than this number, your search results will be truncated. If you send a request with startAt=98 and maxResults=20, it will only return 2 users.\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"permissions\": {\n      \"description\": \"comma separated list of permissions for project or issue returned users must have\",\n      \"type\": \"string\"\n    },\n    \"projectKey\": {\n      \"description\": \"the optional project key to search for users with if no issueKey is supplied.\",\n      \"type\": \"string\"\n    },\n    \"startAt\": {\n      \"description\": \"The index of the first user to return (0-based). Please note that the startAt parameter will be deprecated in a future release of Jira 10.3.x\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"username\": {\n      \"description\": \"the username filter, list includes all users if unspecified\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the FindUsersWithAllPermissions tool (Status: 200, Content-Type: application/json)
const FindUsersWithAllPermissionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of users that match the search string.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **deleted** (Type: boolean):\n      - Example: 'false'\n  - **name** (Type: string):\n      - Example: 'fred'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2.0/user?username=fred'\n  - **active** (Type: boolean):\n      - Example: 'true'\n  - **emailAddress** (Type: string):\n      - Example: 'fred@example.com'\n  - **groups** (Type: object):\n    - **pagingCallback** (Type: object):\n    - **size** (Type: integer, int32):\n    - **[cyclic reference]**\n    - **maxResults** (Type: integer, int32):\n  - **lastLoginTime** (Type: string):\n      - Example: '2023-08-30T16:37:01+1000'\n  - **expand** (Type: string):\n  - **timeZone** (Type: string):\n      - Example: 'Australia/Sydney'\n  - **displayName** (Type: string):\n      - Example: 'Fred F. User'\n  - **key** (Type: string):\n      - Example: 'JIRAUSER10100'\n  - **locale** (Type: string):\n      - Example: 'en_AU'\n  - **applicationRoles** (Type: object):\n      - Example: '[\"jira-core\",\"jira-admin\",\"important\"]'\n    - **callback** (Type: object):\n    - **maxResults** (Type: integer, int32):\n    - **[cyclic reference]**\n    - **size** (Type: integer, int32):\n  - **avatarUrls** (Type: object):\n      - Example: '{\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\",\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\"}'\n    - **Additional Properties**:\n      - **property value** (Type: string, uri):\n          - Example: '{\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large&ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small&ownerId=fred\",\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall&ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium&ownerId=fred\"}'\n"

// NewFindUsersWithAllPermissionsMCPTool creates the MCP Tool instance for FindUsersWithAllPermissions
func NewFindUsersWithAllPermissionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"FindUsersWithAllPermissions",
		"Find users with all specified permissions - Returns a list of active users that match the search string and have all specified permissions for the project or issue. This resource can be accessed by users with ADMINISTER_PROJECT permission for the project or global ADMIN or SYSADMIN rights. This endpoint can cause serious performance issues and will be removed in Jira 9.0.",
		[]byte(FindUsersWithAllPermissionsInputSchema),
	)
}

// FindUsersWithAllPermissionsHandler is the handler function for the FindUsersWithAllPermissions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func FindUsersWithAllPermissionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/permission/search", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "FindUsersWithAllPermissions"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
