package mcptools

import (
	"confluence-mcp/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the GetPermissionsGrantedToGroup1 tool
const GetPermissionsGrantedToGroup1InputSchema = "{\n  \"properties\": {\n    \"groupName\": {\n      \"type\": \"string\"\n    },\n    \"spaceKey\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"groupName\",\n    \"spaceKey\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetPermissionsGrantedToGroup1 tool (Status: 200, Content-Type: application/json)
const GetPermissionsGrantedToGroup1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of the space permissions granted to the group.\n\n## Response Structure\n\n- Structure (Type: array):\n  - **Items** (Type: object):\n    - **subject** (Type: unknown):\n    - **operation** (Type: object):\n      - **targetType** (Type: string):\n          - Example: 'space'\n      - **operationKey** (Type: string):\n          - Example: 'read'\n    - **spaceId** (Type: integer, int64):\n    - **spaceKey** (Type: string):\n"

// Response Template for the GetPermissionsGrantedToGroup1 tool (Status: 401, Content-Type: application/json)
const GetPermissionsGrantedToGroup1ResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the calling User is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetPermissionsGrantedToGroup1 tool (Status: 403, Content-Type: application/json)
const GetPermissionsGrantedToGroup1ResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling User does not have necessary permission.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the GetPermissionsGrantedToGroup1 tool (Status: 404, Content-Type: application/json)
const GetPermissionsGrantedToGroup1ResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if group with specified name not found.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewGetPermissionsGrantedToGroup1MCPTool creates the MCP Tool instance for GetPermissionsGrantedToGroup1
func NewGetPermissionsGrantedToGroup1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPermissionsGrantedToGroup1",
		"Gets the permissions granted to a group in a space - Returns list of permissions granted to a group in the particular space.\n\nExample request URI's:\n"+"\x60"+"https://example.com/confluence/rest/api/space/TESTSPACE/permissions/group/test-group-name"+"\x60"+"",
		[]byte(GetPermissionsGrantedToGroup1InputSchema),
	)
}

// GetPermissionsGrantedToGroup1Handler is the handler function for the GetPermissionsGrantedToGroup1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPermissionsGrantedToGroup1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/confluence/rest/api/space/{spaceKey}/permissions/group/{groupName}", args, []string{"groupName", "spaceKey"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPermissionsGrantedToGroup1")
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
