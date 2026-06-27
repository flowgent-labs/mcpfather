package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetFavouriteFilters tool
const GetFavouriteFiltersInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetFavouriteFilters tool (Status: 200, Content-Type: application/json)
const GetFavouriteFiltersResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of favourite filters\n\n## Response Structure\n\n- Structure (Type: object):\n  - **sharedUsers** (Type: object):\n    - **pagingCallback** (Type: object):\n    - **size** (Type: integer, int32):\n        - Example: '50'\n    - **backingListSize** (Type: integer, int32):\n    - **[cyclic reference]**\n    - **items** (Type: array):\n        - Example: '[]'\n      - **Items** (Type: object):\n        - **applicationRoles** (Type: object):\n            - Example: '[\"jira-core\",\"jira-admin\",\"important\"]'\n          - **callback** (Type: object):\n          - **maxResults** (Type: integer, int32):\n          - **[cyclic reference]**\n          - **size** (Type: integer, int32):\n        - **avatarUrls** (Type: object):\n            - Example: '{\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall\\u0026ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small\\u0026ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium\\u0026ownerId=fred\",\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large\\u0026ownerId=fred\"}'\n          - **Additional Properties**:\n            - **property value** (Type: string, uri):\n                - Example: '{\"48x48\":\"http://www.example.com/jira/secure/useravatar?size=large&ownerId=fred\",\"24x24\":\"http://www.example.com/jira/secure/useravatar?size=small&ownerId=fred\",\"16x16\":\"http://www.example.com/jira/secure/useravatar?size=xsmall&ownerId=fred\",\"32x32\":\"http://www.example.com/jira/secure/useravatar?size=medium&ownerId=fred\"}'\n        - **deleted** (Type: boolean):\n            - Example: 'false'\n        - **name** (Type: string):\n            - Example: 'fred'\n        - **self** (Type: string, uri):\n            - Example: 'http://www.example.com/jira/rest/api/2.0/user?username=fred'\n        - **active** (Type: boolean):\n            - Example: 'true'\n        - **emailAddress** (Type: string):\n            - Example: 'fred@example.com'\n        - **groups** (Type: object):\n          - **size** (Type: integer, int32):\n          - **callback** (Type: object):\n          - **maxResults** (Type: integer, int32):\n          - **[cyclic reference]**\n        - **lastLoginTime** (Type: string):\n            - Example: '2023-08-30T16:37:01+1000'\n        - **expand** (Type: string):\n        - **timeZone** (Type: string):\n            - Example: 'Australia/Sydney'\n        - **displayName** (Type: string):\n            - Example: 'Fred F. User'\n        - **key** (Type: string):\n            - Example: 'JIRAUSER10100'\n        - **locale** (Type: string):\n            - Example: 'en_AU'\n    - **maxResults** (Type: integer, int32):\n        - Example: '50'\n  - **viewUrl** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/issues/?filter=10000'\n  - **editable** (Type: boolean):\n      - Example: 'false'\n  - **id** (Type: string):\n      - Example: '10000'\n  - **jql** (Type: string):\n      - Example: 'type = Bug and resolution is empty'\n  - **searchUrl** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/search?jql=type%20%3D%20Bug%20and%20resolutino%20is%20empty'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/filter/10000'\n  - **sharePermissions** (Type: array):\n      - Example: '[]'\n    - **Items** (Type: object):\n      - **edit** (Type: boolean):\n          - Example: 'false'\n      - **group** (Type: object):\n        - **self** (Type: string, uri):\n            - Example: 'http://www.example.com/jira/rest/api/2/group?groupname=jira-administrators'\n        - **name** (Type: string):\n            - Example: 'jira-administrators'\n      - **id** (Type: integer, int64):\n          - Example: '10000'\n      - **project** (Type: object):\n        - **self** (Type: string, uri):\n            - Example: 'http://www.example.com/jira/rest/api/2/project/EX'\n        - **archived** (Type: boolean):\n            - Example: 'false'\n        - **avatarUrls** (Type: object):\n            - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n          - **Additional Properties**:\n            - **property value** (Type: string):\n                - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n        - **description** (Type: string):\n            - Example: 'Example'\n        - **id** (Type: string):\n            - Example: '10000'\n        - **key** (Type: string):\n            - Example: 'EX'\n        - **name** (Type: string):\n            - Example: 'Example'\n      - **role** (Type: object):\n        - **id** (Type: integer, int64):\n            - Example: '10360'\n        - **name** (Type: string):\n            - Example: 'Developers'\n        - **self** (Type: string, uri):\n            - Example: 'http://www.example.com/jira/rest/api/2/project/MKY/role/10360'\n        - **actors** (Type: array):\n          - **Items** (Type: object):\n            - **avatarUrl** (Type: string, uri):\n            - **name** (Type: string):\n                - Example: 'jira-developers'\n        - **description** (Type: string):\n            - Example: 'A project role that represents developers in a project'\n      - **type** (Type: string):\n          - Example: 'global'\n      - **[cyclic reference]**\n      - **view** (Type: boolean):\n          - Example: 'true'\n  - **name** (Type: string):\n      - Example: 'All Open Bugs'\n  - **[cyclic reference]**\n  - **description** (Type: string):\n      - Example: 'Lists all open bugs'\n  - **favourite** (Type: boolean):\n      - Example: 'true'\n"

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetFavouriteFilters"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
