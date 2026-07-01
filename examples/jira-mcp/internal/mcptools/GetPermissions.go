package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetPermissions tool
const GetPermissionsInputSchema = "{\n  \"properties\": {\n    \"issueId\": {\n      \"description\": \"id of the issue to scope returned permissions for.\",\n      \"type\": \"string\"\n    },\n    \"issueKey\": {\n      \"description\": \"key of the issue to scope returned permissions for.\",\n      \"type\": \"string\"\n    },\n    \"projectId\": {\n      \"description\": \"id of project to scope returned permissions for.\",\n      \"type\": \"string\"\n    },\n    \"projectKey\": {\n      \"description\": \"key of project to scope returned permissions for.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the GetPermissions tool (Status: 200, Content-Type: application/json)
const GetPermissionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of all permissions in Jira and whether the user has them.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **permissions**: A map of permission keys to permission objects. (Type: object):\n      - Example: '{\"EDIT_ISSUE\":{\"description\":\"Ability to edit issues.\",\"havePermission\":true,\"id\":\"EDIT_ISSUE\",\"name\":\"Edit Issues\",\"type\":\"USER\"}}'\n    - **Additional Properties**:\n      - **property value**: A map of permission keys to permission objects. (Type: object):\n          - Example: '{\"EDIT_ISSUE\":{\"description\":\"Ability to edit issues.\",\"havePermission\":true,\"id\":\"EDIT_ISSUE\",\"name\":\"Edit Issues\",\"type\":\"USER\"}}'\n        - **name** (Type: string):\n        - **type** (Type: string):\n            - Enum: ['GLOBAL', 'PROJECT']\n        - **description** (Type: string):\n        - **id** (Type: string):\n        - **key** (Type: string):\n"

// NewGetPermissionsMCPTool creates the MCP Tool instance for GetPermissions
func NewGetPermissionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPermissions",
		"Get permissions for the logged in user - Returns all permissions in the system and whether the currently logged in user has them. You can optionally provide a specific context to get permissions for (projectKey OR projectId OR issueKey OR issueId)",
		[]byte(GetPermissionsInputSchema),
	)
}

// GetPermissionsHandler is the handler function for the GetPermissions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPermissionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/mypermissions", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPermissions")
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
