package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the AssignPermissionScheme tool
const AssignPermissionSchemeInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Object that contains an id of the scheme\",\n      \"properties\": {\n        \"id\": {\n          \"example\": 10000,\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"expand\": {\n      \"description\": \"Use expand to include additional information about permission schemes in the response. This parameter accepts a comma-separated list of expandable options. Expand options include: all and field.\",\n      \"type\": \"string\"\n    },\n    \"projectKeyOrId\": {\n      \"description\": \"Key or id of the project\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"projectKeyOrId\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the AssignPermissionScheme tool (Status: 200, Content-Type: application/json)
const AssignPermissionSchemeResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Shortened details of the newly associated permission scheme.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **self** (Type: string, uri):\n      - Example: 'http://www.example.com/jira/rest/api/2/permissionscheme/10100'\n  - **description** (Type: string):\n      - Example: 'description'\n  - **expand** (Type: string):\n  - **id** (Type: integer, int64):\n      - Example: '10100'\n  - **name** (Type: string):\n      - Example: 'permission scheme name'\n  - **permissions** (Type: array):\n    - **Items** (Type: object):\n      - **holder** (Type: object):\n        - **field** (Type: object):\n          - **custom** (Type: boolean):\n              - Example: 'false'\n          - **id** (Type: string):\n              - Example: 'description'\n          - **name** (Type: string):\n              - Example: 'Description'\n          - **navigable** (Type: boolean):\n              - Example: 'true'\n          - **orderable** (Type: boolean):\n              - Example: 'true'\n          - **schema** (Type: object):\n              - Example: '{}'\n            - **custom** (Type: string):\n                - Example: 'null'\n            - **customId** (Type: integer, int64):\n            - **items** (Type: string):\n                - Example: 'null'\n            - **system** (Type: string):\n                - Example: 'summary'\n            - **type** (Type: string):\n                - Example: 'string'\n          - **searchable** (Type: boolean):\n              - Example: 'true'\n          - **clauseNames** (Type: array):\n              - Unique Items: true\n              - Example: '\"[description]\"'\n            - **Items** (Type: string):\n                - Example: '[description]'\n        - **group** (Type: object):\n          - **self** (Type: string, uri):\n              - Example: 'http://www.example.com/jira/rest/api/2/group?groupname=jira-administrators'\n          - **name** (Type: string):\n              - Example: 'jira-administrators'\n        - **parameter** (Type: string):\n            - Example: 'admin'\n        - **projectRole** (Type: object):\n          - **actors** (Type: array):\n            - **Items** (Type: object):\n              - **avatarUrl** (Type: string, uri):\n              - **name** (Type: string):\n                  - Example: 'jira-developers'\n          - **description** (Type: string):\n              - Example: 'A project role that represents developers in a project'\n          - **id** (Type: integer, int64):\n              - Example: '10360'\n          - **name** (Type: string):\n              - Example: 'Developers'\n          - **self** (Type: string, uri):\n              - Example: 'http://www.example.com/jira/rest/api/2/project/MKY/role/10360'\n        - **type** (Type: string):\n            - Example: 'user'\n        - **user** (Type: object):\n          - **displayName** (Type: string):\n              - Example: 'Fred F. User'\n          - **emailAddress** (Type: string):\n              - Example: 'fred@example.com'\n          - **key** (Type: string):\n              - Example: 'fred'\n          - **name** (Type: string):\n              - Example: 'Fred'\n          - **self** (Type: string):\n              - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n          - **timeZone** (Type: string):\n              - Example: 'Australia/Sydney'\n          - **active** (Type: boolean):\n              - Example: 'true'\n          - **avatarUrls** (Type: object):\n              - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n            - **Additional Properties**:\n              - **property value** (Type: string):\n                  - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n      - **id** (Type: integer, int64):\n          - Example: '10100'\n      - **permission** (Type: string):\n          - Example: 'permission scheme name'\n      - **self** (Type: string, uri):\n          - Example: 'http://www.example.com/jira/rest/api/2/permissionscheme/10100'\n"

// NewAssignPermissionSchemeMCPTool creates the MCP Tool instance for AssignPermissionScheme
func NewAssignPermissionSchemeMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"AssignPermissionScheme",
		"Assign permission scheme to project - Assigns a permission scheme with a project",
		[]byte(AssignPermissionSchemeInputSchema),
	)
}

// AssignPermissionSchemeHandler is the handler function for the AssignPermissionScheme tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func AssignPermissionSchemeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/project/{projectKeyOrId}/permissionscheme", args, []string{"projectKeyOrId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "AssignPermissionScheme")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
