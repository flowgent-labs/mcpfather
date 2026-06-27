package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetAllSpacePermissions tool
const GetAllSpacePermissionsInputSchema = "{\n  \"properties\": {\n    \"spaceKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetAllSpacePermissions tool (Status: 200, Content-Type: application/json)
const GetAllSpacePermissionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of all permissions granted to users/groups/anonymous user in specified space.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **operation** (Type: object):\n      - **targetType** (Type: string):\n          - Example: 'space'\n      - **operationKey** (Type: string):\n          - Example: 'read'\n    - **spaceId** (Type: integer, int64):\n    - **spaceKey** (Type: string):\n    - **subject** (Type: unknown):\n"

// Response Template for the GetAllSpacePermissions tool (Status: 401, Content-Type: application/json)
const GetAllSpacePermissionsResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the calling User is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetAllSpacePermissions tool (Status: 403, Content-Type: application/json)
const GetAllSpacePermissionsResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling User does not have necessary permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetAllSpacePermissionsMCPTool creates the MCP Tool instance for GetAllSpacePermissions
func NewGetAllSpacePermissionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetAllSpacePermissions",
		"Get all space permissions - Returns list of permissions granted to users and groups in the particular space.\n\nExample request URI's:\n"+"\x60"+"https://example.com/confluence/rest/api/space/TESTSPACE/permissions"+"\x60"+"",
		[]byte(GetAllSpacePermissionsInputSchema),
	)
}

// GetAllSpacePermissionsHandler is the handler function for the GetAllSpacePermissions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetAllSpacePermissionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/space/{spaceKey}/permissions", args, []string{"spaceKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetAllSpacePermissions"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
