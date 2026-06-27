package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetPermissionsGrantedToAnonymousUsers tool
const GetPermissionsGrantedToAnonymousUsersInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetPermissionsGrantedToAnonymousUsers tool (Status: 200, Content-Type: application/json)
const GetPermissionsGrantedToAnonymousUsersResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of the permissions granted to the group.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **subject** (Type: unknown):\n    - **operation** (Type: object):\n      - **operationKey** (Type: string):\n          - Example: 'read'\n      - **targetType** (Type: string):\n          - Example: 'space'\n"

// Response Template for the GetPermissionsGrantedToAnonymousUsers tool (Status: 401, Content-Type: application/json)
const GetPermissionsGrantedToAnonymousUsersResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the calling User is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetPermissionsGrantedToAnonymousUsers tool (Status: 403, Content-Type: application/json)
const GetPermissionsGrantedToAnonymousUsersResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling User does not have necessary permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetPermissionsGrantedToAnonymousUsersMCPTool creates the MCP Tool instance for GetPermissionsGrantedToAnonymousUsers
func NewGetPermissionsGrantedToAnonymousUsersMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPermissionsGrantedToAnonymousUsers",
		"Gets the permissions granted to an anonymous user - Returns list of permissions granted to anonymous user.\n\nExample request URI's:\n\n       "+"\x60"+"https://example.com/confluence/rest/api/permissions/anonymous"+"\x60"+"",
		[]byte(GetPermissionsGrantedToAnonymousUsersInputSchema),
	)
}

// GetPermissionsGrantedToAnonymousUsersHandler is the handler function for the GetPermissionsGrantedToAnonymousUsers tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPermissionsGrantedToAnonymousUsersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/permissions/anonymous", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "GetPermissionsGrantedToAnonymousUsers"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
