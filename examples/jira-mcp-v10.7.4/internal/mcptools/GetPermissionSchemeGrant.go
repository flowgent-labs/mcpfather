package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the GetPermissionSchemeGrant tool
const GetPermissionSchemeGrantInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"Use expand to include full beans in the response. This parameter accepts a comma-separated list of expandable elements. Use 'permissions' to include permissions in the response.\",\n      \"type\": \"string\"\n    },\n    \"permissionId\": {\n      \"description\": \"The id of the permission grant.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    },\n    \"schemeId\": {\n      \"description\": \"The id of the permission scheme.\",\n      \"format\": \"int64\",\n      \"type\": \"integer\"\n    }\n  },\n  \"required\": [\n    \"permissionId\",\n    \"schemeId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPermissionSchemeGrant tool (Status: 200, Content-Type: application/json)
const GetPermissionSchemeGrantResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Permission grant\n\n## Response Structure\n\n- Structure (Type: object):\n  - **holder** (Type: object):\n    - **parameter** (Type: string):\n        - Example: 'admin'\n    - **projectRole** (Type: object):\n      - **id** (Type: integer, int64):\n          - Example: '10360'\n      - **name** (Type: string):\n          - Example: 'Developers'\n      - **self** (Type: string, uri):\n          - Example: 'http://www.example.com/jira/rest/api/2/project/MKY/role/10360'\n      - **actors** (Type: array):\n        - **Items** (Type: object):\n          - **avatarUrl** (Type: string, uri):\n          - **name** (Type: string):\n              - Example: 'jira-developers'\n      - **description** (Type: string):\n          - Example: 'A project role that represents developers in a project'\n    - **type** (Type: string):\n        - Example: 'user'\n    - **user** (Type: object):\n      - **active** (Type: boolean):\n          - Example: 'true'\n      - **avatarUrls** (Type: object):\n          - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n        - **Additional Properties**:\n          - **property value** (Type: string):\n              - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n      - **displayName** (Type: string):\n          - Example: 'Fred F. User'\n      - **emailAddress** (Type: string):\n          - Example: 'fred@example.com'\n      - **key** (Type: string):\n          - Example: 'fred'\n      - **name** (Type: string):\n          - Example: 'Fred'\n      - **self** (Type: string):\n          - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n      - **timeZone** (Type: string):\n          - Example: 'Australia/Sydney'\n    - **field** (Type: object):\n      - **orderable** (Type: boolean):\n          - Example: 'true'\n      - **schema** (Type: object):\n          - Example: '{}'\n        - **system** (Type: string):\n            - Example: 'summary'\n        - **type** (Type: string):\n            - Example: 'string'\n        - **custom** (Type: string):\n            - Example: 'null'\n        - **customId** (Type: integer, int64):\n        - **items** (Type: string):\n            - Example: 'null'\n      - **searchable** (Type: boolean):\n          - Example: 'true'\n      - **clauseNames** (Type: array):\n          - Unique Items: true\n          - Example: '\"[description]\"'\n        - **Items** (Type: string):\n            - Example: '[description]'\n      - **custom** (Type: boolean):\n          - Example: 'false'\n      - **id** (Type: string):\n          - Example: 'description'\n      - **name** (Type: string):\n          - Example: 'Description'\n      - **navigable** (Type: boolean):\n          - Example: 'true'\n    - **group** (Type: object):\n      - **name** (Type: string):\n          - Example: 'jira-administrators'\n      - **self** (Type: string, uri):\n          - Example: 'http://www.example.com/jira/rest/api/2/group?groupname=jira-administrators'\n  - **id** (Type: integer, int64):\n      - Example: '10100'\n  - **permission** (Type: string):\n      - Example: 'permission scheme name'\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/permissionscheme/10100'\n"

// NewGetPermissionSchemeGrantMCPTool creates the MCP Tool instance for GetPermissionSchemeGrant
func NewGetPermissionSchemeGrantMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPermissionSchemeGrant",
		"Get a permission grant by ID - Returns a permission grant identified by the given id.",
		[]byte(GetPermissionSchemeGrantInputSchema),
	)
}

// GetPermissionSchemeGrantHandler is the handler function for the GetPermissionSchemeGrant tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPermissionSchemeGrantHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/permissionscheme/{schemeId}/permission/{permissionId}", args, []string{"permissionId", "schemeId"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetPermissionSchemeGrant"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
