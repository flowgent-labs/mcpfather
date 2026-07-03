package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetDuplicatedUsersCount tool
const GetDuplicatedUsersCountInputSchema = "{\n  \"properties\": {\n    \"flush\": {\n      \"description\": \"if set to true forces cache flush, user must be sysadmin for this parameter to have an effect.\",\n      \"type\": \"boolean\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetDuplicatedUsersCount tool (Status: 200, Content-Type: application/json)
const GetDuplicatedUsersCountResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of users that match the search string.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **locale** (Type: string):\n      - Example: 'en_AU'\n  - **key** (Type: string):\n      - Example: 'JIRAUSER10100'\n  - **name** (Type: string):\n      - Example: 'fred'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2.0/user?username=fred'\n  - **timeZone** (Type: string):\n      - Example: 'Australia/Sydney'\n  - **lastLoginTime** (Type: string):\n      - Example: '2023-08-30T16:37:01+1000'\n  - **deleted** (Type: boolean):\n      - Example: 'false'\n  - **groups** (Type: object):\n    - **callback** (Type: object):\n    - **maxResults** (Type: integer, int32):\n    - **[cyclic reference]**\n    - **size** (Type: integer, int32):\n  - **active** (Type: boolean):\n      - Example: 'true'\n  - **applicationRoles** (Type: object):\n      - Example: '[\"jira-core\",\"jira-admin\",\"important\"]'\n    - **maxResults** (Type: integer, int32):\n    - **pagingCallback** (Type: object):\n    - **size** (Type: integer, int32):\n    - **[cyclic reference]**\n  - **avatarUrls** (Type: object):\n      - Example: '{\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\",\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\"}'\n    - **Additional Properties**:\n      - **property value** (Type: string, uri):\n          - Example: '{\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large&ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small&ownerId=fred\",\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall&ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium&ownerId=fred\"}'\n  - **displayName** (Type: string):\n      - Example: 'Fred F. User'\n  - **emailAddress** (Type: string):\n      - Example: 'fred@example.com'\n  - **expand** (Type: string):\n"

// NewGetDuplicatedUsersCountMCPTool creates the MCP Tool instance for GetDuplicatedUsersCount
func NewGetDuplicatedUsersCountMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetDuplicatedUsersCount",
		"Get duplicated users count - Returns a list of users that match the search string. This resource cannot be accessed anonymously.\nDuplicated means that the user has an account in more than one directory\nand either more than one account is active or the only active account does not belong to the directory\nwith the highest priority.\nThe data returned by this endpoint is cached for 10 minutes and the cache is flushed when any User Directory\nis added, removed, enabled, disabled, or synchronized.\nA System Administrator can also flush the cache manually.\nRelated JAC ticket: https://jira.atlassian.com/browse/JRASERVER-68797",
		[]byte(GetDuplicatedUsersCountInputSchema),
	)
}

// GetDuplicatedUsersCountHandler is the handler function for the GetDuplicatedUsersCount tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetDuplicatedUsersCountHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/duplicated/count", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetDuplicatedUsersCount")
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
