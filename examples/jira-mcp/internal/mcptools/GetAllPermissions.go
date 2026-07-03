package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetAllPermissions tool
const GetAllPermissionsInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetAllPermissions tool (Status: 200, Content-Type: application/json)
const GetAllPermissionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a list of all permissions in Jira.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **permissions**: A map of permission keys to permission objects. (Type: object):\n      - Example: '{\"EDIT_ISSUE\":{\"description\":\"Ability to edit issues.\",\"havePermission\":true,\"id\":\"EDIT_ISSUE\",\"name\":\"Edit Issues\",\"type\":\"USER\"}}'\n    - **Additional Properties**:\n      - **property value**: A map of permission keys to permission objects. (Type: object):\n          - Example: '{\"EDIT_ISSUE\":{\"description\":\"Ability to edit issues.\",\"havePermission\":true,\"id\":\"EDIT_ISSUE\",\"name\":\"Edit Issues\",\"type\":\"USER\"}}'\n        - **key** (Type: string):\n        - **name** (Type: string):\n        - **type** (Type: string):\n            - Enum: ['GLOBAL', 'PROJECT']\n        - **description** (Type: string):\n        - **id** (Type: string):\n"

// NewGetAllPermissionsMCPTool creates the MCP Tool instance for GetAllPermissions
func NewGetAllPermissionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllPermissions",
		"Get all permissions present in Jira instance - Returns all permissions that are present in the Jira instance - Global, Project and the global ones added by plugins",
		[]byte(GetAllPermissionsInputSchema),
	)
}

// GetAllPermissionsHandler is the handler function for the GetAllPermissions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllPermissionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/permissions", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetAllPermissions")
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
