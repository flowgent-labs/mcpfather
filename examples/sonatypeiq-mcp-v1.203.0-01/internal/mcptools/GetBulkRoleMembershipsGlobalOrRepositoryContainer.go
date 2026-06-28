package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetBulkRoleMembershipsGlobalOrRepositoryContainer tool
const GetBulkRoleMembershipsGlobalOrRepositoryContainerInputSchema = "{\n  \"properties\": {\n    \"ownerType\": {\n      \"description\": \"Enter the value for ownerType.\",\n      \"enum\": [\n        \"repository_container\",\n        \"global\"\n      ],\n      \"pattern\": \"global|repository_container\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetBulkRoleMembershipsGlobalOrRepositoryContainer tool (Status: 200, Content-Type: application/json)
const GetBulkRoleMembershipsGlobalOrRepositoryContainerResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains all roles with their members. Also includes a flag indicating whether group search is enabled.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **groupSearchEnabled** (Type: boolean):\n  - **membersByRole** (Type: array):\n    - **Items** (Type: object):\n      - **roleDescription** (Type: string):\n      - **roleId** (Type: string):\n      - **roleName** (Type: string):\n      - **membersByOwner** (Type: array):\n        - **Items** (Type: object):\n          - **ownerId** (Type: string):\n          - **ownerName** (Type: string):\n          - **ownerType** (Type: string):\n          - **members** (Type: array):\n            - **Items** (Type: object):\n              - **realm** (Type: string):\n              - **type** (Type: string):\n                  - Enum: ['USER', 'GROUP']\n              - **displayName** (Type: string):\n              - **email** (Type: string):\n              - **internalName** (Type: string):\n"

// NewGetBulkRoleMembershipsGlobalOrRepositoryContainerMCPTool creates the MCP Tool instance for GetBulkRoleMembershipsGlobalOrRepositoryContainer
func NewGetBulkRoleMembershipsGlobalOrRepositoryContainerMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetBulkRoleMembershipsGlobalOrRepositoryContainer",
		"Use this method to retrieve all role memberships for global or repository container context with full details including role names, descriptions, and member information.\n\nPermissions required: Edit system configuration and users for a global context or view IQ elements for a non-global context",
		[]byte(GetBulkRoleMembershipsGlobalOrRepositoryContainerInputSchema),
	)
}

// GetBulkRoleMembershipsGlobalOrRepositoryContainerHandler is the handler function for the GetBulkRoleMembershipsGlobalOrRepositoryContainer tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetBulkRoleMembershipsGlobalOrRepositoryContainerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/roleMemberships/{ownerType}/roles", args, []string{"ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetBulkRoleMembershipsGlobalOrRepositoryContainer")
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
