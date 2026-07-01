package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetFavouriteFilters tool
const GetFavouriteFiltersInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetFavouriteFilters tool (Status: 200, Content-Type: application/json)
const GetFavouriteFiltersResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of favourite filters\n\n## Response Structure\n\n- Structure (Type: object):\n  - **description** (Type: string):\n      - Example: 'Lists all open bugs'\n  - **sharePermissions** (Type: array):\n      - Example: '[]'\n    - **Items** (Type: object):\n      - **view** (Type: boolean):\n          - Example: 'true'\n      - **edit** (Type: boolean):\n          - Example: 'false'\n      - **group** (Type: object):\n        - **name** (Type: string):\n            - Example: 'jira-administrators'\n        - **self** (Type: string, uri):\n            - Example: 'http://www.example.com/jira/rest/api/2/group?groupname=jira-administrators'\n      - **id** (Type: integer, int64):\n          - Example: '10000'\n      - **project** (Type: object):\n        - **description** (Type: string):\n            - Example: 'Example'\n        - **id** (Type: string):\n            - Example: '10000'\n        - **key** (Type: string):\n            - Example: 'EX'\n        - **name** (Type: string):\n            - Example: 'Example'\n        - **self** (Type: string, uri):\n            - Example: 'http://www.example.com/jira/rest/api/2/project/EX'\n        - **archived** (Type: boolean):\n            - Example: 'false'\n        - **avatarUrls** (Type: object):\n            - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n          - **Additional Properties**:\n            - **property value** (Type: string):\n                - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n      - **role** (Type: object):\n        - **actors** (Type: array):\n          - **Items** (Type: object):\n            - **avatarUrl** (Type: string, uri):\n            - **name** (Type: string):\n                - Example: 'jira-developers'\n        - **description** (Type: string):\n            - Example: 'A project role that represents developers in a project'\n        - **id** (Type: integer, int64):\n            - Example: '10360'\n        - **name** (Type: string):\n            - Example: 'Developers'\n        - **self** (Type: string, uri):\n            - Example: 'http://www.example.com/jira/rest/api/2/project/MKY/role/10360'\n      - **type** (Type: string):\n          - Example: 'global'\n      - **user** (Type: object):\n        - **self** (Type: string, uri):\n            - Example: 'http://www.example.com/jira/rest/api/2.0/user?username=fred'\n        - **timeZone** (Type: string):\n            - Example: 'Australia/Sydney'\n        - **active** (Type: boolean):\n            - Example: 'true'\n        - **deleted** (Type: boolean):\n            - Example: 'false'\n        - **locale** (Type: string):\n            - Example: 'en_AU'\n        - **avatarUrls** (Type: object):\n            - Example: '{\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\",\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\"}'\n          - **Additional Properties**:\n            - **property value** (Type: string, uri):\n                - Example: '{\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large&ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small&ownerId=fred\",\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall&ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium&ownerId=fred\"}'\n        - **displayName** (Type: string):\n            - Example: 'Fred F. User'\n        - **key** (Type: string):\n            - Example: 'JIRAUSER10100'\n        - **applicationRoles** (Type: object):\n            - Example: '[\"jira-core\",\"jira-admin\",\"important\"]'\n          - **pagingCallback** (Type: object):\n          - **size** (Type: integer, int32):\n          - **[cyclic reference]**\n          - **maxResults** (Type: integer, int32):\n        - **emailAddress** (Type: string):\n            - Example: 'fred@example.com'\n        - **expand** (Type: string):\n        - **groups** (Type: object):\n          - **maxResults** (Type: integer, int32):\n          - **pagingCallback** (Type: object):\n          - **size** (Type: integer, int32):\n          - **[cyclic reference]**\n        - **name** (Type: string):\n            - Example: 'fred'\n        - **lastLoginTime** (Type: string):\n            - Example: '2023-08-30T16:37:01+1000'\n  - **favourite** (Type: boolean):\n      - Example: 'true'\n  - **[cyclic reference]**\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/filter/10000'\n  - **sharedUsers** (Type: object):\n    - **items** (Type: array):\n        - Example: '[]'\n      - **[cyclic reference]**\n    - **maxResults** (Type: integer, int32):\n        - Example: '50'\n    - **pagingCallback** (Type: object):\n    - **size** (Type: integer, int32):\n        - Example: '50'\n    - **backingListSize** (Type: integer, int32):\n    - **[cyclic reference]**\n  - **jql** (Type: string):\n      - Example: 'type = Bug and resolution is empty'\n  - **name** (Type: string):\n      - Example: 'All Open Bugs'\n  - **viewUrl** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/issues/?filter=10000'\n  - **editable** (Type: boolean):\n      - Example: 'false'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **searchUrl** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/search?jql=type%20%3D%20Bug%20and%20resolutino%20is%20empty'\n"

// NewGetFavouriteFiltersMCPTool creates the MCP Tool instance for GetFavouriteFilters
func NewGetFavouriteFiltersMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetFavouriteFilters",
		"Get favourite filters - Returns the favourite filters of the logged-in user",
		[]byte(GetFavouriteFiltersInputSchema),
	)
}

// GetFavouriteFiltersHandler is the handler function for the GetFavouriteFilters tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetFavouriteFiltersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/filter/favourite", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetFavouriteFilters")
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
