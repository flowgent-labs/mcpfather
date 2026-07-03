package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetPermissionsGrantedToUser tool
const GetPermissionsGrantedToUserInputSchema = "{\n  \"properties\": {\n    \"user\": {\n      \"description\": \"the key or username of the user to look up.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"user\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPermissionsGrantedToUser tool (Status: 200, Content-Type: application/json)
const GetPermissionsGrantedToUserResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of the permissions granted to the user.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **subject** (Type: unknown):\n    - **operation** (Type: object):\n      - **operationKey** (Type: string):\n          - Example: 'read'\n      - **targetType** (Type: string):\n          - Example: 'space'\n"

// Response Template for the GetPermissionsGrantedToUser tool (Status: 401, Content-Type: application/json)
const GetPermissionsGrantedToUserResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the calling User is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetPermissionsGrantedToUser tool (Status: 403, Content-Type: application/json)
const GetPermissionsGrantedToUserResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling User does not have necessary permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetPermissionsGrantedToUser tool (Status: 404, Content-Type: application/json)
const GetPermissionsGrantedToUserResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if the user with specified key not found.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetPermissionsGrantedToUserMCPTool creates the MCP Tool instance for GetPermissionsGrantedToUser
func NewGetPermissionsGrantedToUserMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPermissionsGrantedToUser",
		"Gets global permissions granted to a user - Returns list of permissions granted to user.\n\nExample request URI's:\n\n       with userKey: "+"\x60"+"https://example.com/confluence/rest/api/permissions/user/{userKey}"+"\x60"+"\n\n       with username: "+"\x60"+"https://example.com/confluence/rest/api/permissions/user/{username}"+"\x60"+"",
		[]byte(GetPermissionsGrantedToUserInputSchema),
	)
}

// GetPermissionsGrantedToUserHandler is the handler function for the GetPermissionsGrantedToUser tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPermissionsGrantedToUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/permissions/user/{user}", args, []string{"user"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPermissionsGrantedToUser")
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
