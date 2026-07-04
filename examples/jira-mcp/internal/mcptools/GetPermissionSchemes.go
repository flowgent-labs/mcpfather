package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetPermissionSchemes tool
const GetPermissionSchemesInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"Use expand to include full beans in the response. This parameter accepts a comma-separated list of expandable elements. Use 'permissions' to include permissions in the response.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetPermissionSchemes tool (Status: 200, Content-Type: application/json)
const GetPermissionSchemesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> List of all permission schemes\n\n## Response Structure\n\n- Structure (Type: object):\n  - **permissionSchemes** (Type: array):\n    - **Items** (Type: object):\n      - **description** (Type: string):\n          - Example: 'description'\n      - **expand** (Type: string):\n      - **id** (Type: integer, int64):\n          - Example: '10100'\n      - **name** (Type: string):\n          - Example: 'permission scheme name'\n      - **permissions** (Type: array):\n        - **Items** (Type: object):\n          - **permission** (Type: string):\n              - Example: 'permission scheme name'\n          - **self** (Type: string, uri):\n              - Example: 'http://www.example.com/jira/rest/api/2/permissionscheme/10100'\n          - **holder** (Type: object):\n            - **expand** (Type: string):\n            - **field** (Type: object):\n              - **id** (Type: string):\n                  - Example: 'description'\n              - **name** (Type: string):\n                  - Example: 'Description'\n              - **navigable** (Type: boolean):\n                  - Example: 'true'\n              - **orderable** (Type: boolean):\n                  - Example: 'true'\n              - **schema** (Type: object):\n                  - Example: '{}'\n                - **custom** (Type: string):\n                    - Example: 'null'\n                - **customId** (Type: integer, int64):\n                - **items** (Type: string):\n                    - Example: 'null'\n                - **system** (Type: string):\n                    - Example: 'summary'\n                - **type** (Type: string):\n                    - Example: 'string'\n              - **searchable** (Type: boolean):\n                  - Example: 'true'\n              - **clauseNames** (Type: array):\n                  - Unique Items: true\n                  - Example: '\"[description]\"'\n                - **Items** (Type: string):\n                    - Example: '[description]'\n              - **custom** (Type: boolean):\n                  - Example: 'false'\n            - **group** (Type: object):\n              - **name** (Type: string):\n                  - Example: 'jira-administrators'\n              - **self** (Type: string, uri):\n                  - Example: 'http://www.example.com/jira/rest/api/2/group?groupname=jira-administrators'\n            - **parameter** (Type: string):\n                - Example: 'admin'\n            - **projectRole** (Type: object):\n              - **actors** (Type: array):\n                - **Items** (Type: object):\n                  - **avatarUrl** (Type: string, uri):\n                  - **name** (Type: string):\n                      - Example: 'jira-developers'\n              - **description** (Type: string):\n                  - Example: 'A project role that represents developers in a project'\n              - **id** (Type: integer, int64):\n                  - Example: '10360'\n              - **name** (Type: string):\n                  - Example: 'Developers'\n              - **self** (Type: string, uri):\n                  - Example: 'http://www.example.com/jira/rest/api/2/project/MKY/role/10360'\n            - **type** (Type: string):\n                - Example: 'user'\n            - **user** (Type: object):\n              - **self** (Type: string):\n                  - Example: 'http://www.example.com/jira/rest/api/2/user?username=fred'\n              - **timeZone** (Type: string):\n                  - Example: 'Australia/Sydney'\n              - **active** (Type: boolean):\n                  - Example: 'true'\n              - **avatarUrls** (Type: object):\n                  - Example: '\"http://www.example.com/jira/secure/projectavatar?size=xsmall\\u0026pid=10000\"'\n                - **Additional Properties**:\n                  - **property value** (Type: string):\n                      - Example: 'http://www.example.com/jira/secure/projectavatar?size=xsmall&pid=10000'\n              - **displayName** (Type: string):\n                  - Example: 'Fred F. User'\n              - **emailAddress** (Type: string):\n                  - Example: 'fred@example.com'\n              - **key** (Type: string):\n                  - Example: 'fred'\n              - **name** (Type: string):\n                  - Example: 'Fred'\n          - **id** (Type: integer, int64):\n              - Example: '10100'\n      - **self** (Type: string, uri):\n          - Example: 'http://www.example.com/jira/rest/api/2/permissionscheme/10100'\n"

// NewGetPermissionSchemesMCPTool creates the MCP Tool instance for GetPermissionSchemes
func NewGetPermissionSchemesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPermissionSchemes",
		"Get all permission schemes - Returns a list of all permission schemes. By default only shortened beans are returned. If you want to include permissions of all the schemes, then specify the permissions expand parameter. Permissions will be included also if you specify any other expand parameter.",
		[]byte(GetPermissionSchemesInputSchema),
	)
}

// GetPermissionSchemesHandler is the handler function for the GetPermissionSchemes tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPermissionSchemesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/permissionscheme", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPermissionSchemes")
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
