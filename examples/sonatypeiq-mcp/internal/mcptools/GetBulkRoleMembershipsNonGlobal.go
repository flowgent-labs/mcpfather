package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetBulkRoleMembershipsNonGlobal tool
const GetBulkRoleMembershipsNonGlobalInputSchema = "{\n  \"properties\": {\n    \"internalOwnerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above. For applications, use the public ID. For organizations, repositories, and repository managers, use the internal ID.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType for which you want to retrieve role memberships.\",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository_manager\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository_manager|repository\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetBulkRoleMembershipsNonGlobal tool (Status: 200, Content-Type: application/json)
const GetBulkRoleMembershipsNonGlobalResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains all roles with their members organized by owner, including inherited members from parent organizations or repository hierarchies. Also includes a flag indicating whether group search is\nenabled.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **groupSearchEnabled** (Type: boolean):\n  - **membersByRole** (Type: array):\n    - **Items** (Type: object):\n      - **membersByOwner** (Type: array):\n        - **Items** (Type: object):\n          - **ownerName** (Type: string):\n          - **ownerType** (Type: string):\n          - **members** (Type: array):\n            - **Items** (Type: object):\n              - **internalName** (Type: string):\n              - **realm** (Type: string):\n              - **type** (Type: string):\n                  - Enum: ['USER', 'GROUP']\n              - **displayName** (Type: string):\n              - **email** (Type: string):\n          - **ownerId** (Type: string):\n      - **roleDescription** (Type: string):\n      - **roleId** (Type: string):\n      - **roleName** (Type: string):\n"

// NewGetBulkRoleMembershipsNonGlobalMCPTool creates the MCP Tool instance for GetBulkRoleMembershipsNonGlobal
func NewGetBulkRoleMembershipsNonGlobalMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetBulkRoleMembershipsNonGlobal",
		"Use this method to retrieve all role memberships with full details including role names, descriptions, and member information organized by owner (for inheritance display).\n\nPermissions required: Edit system configuration and users for a global context or view IQ elements for a non-global context",
		[]byte(GetBulkRoleMembershipsNonGlobalInputSchema),
	)
}

// GetBulkRoleMembershipsNonGlobalHandler is the handler function for the GetBulkRoleMembershipsNonGlobal tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetBulkRoleMembershipsNonGlobalHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/roleMemberships/{ownerType}/{internalOwnerId}/roles", args, []string{"internalOwnerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetBulkRoleMembershipsNonGlobal")
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
