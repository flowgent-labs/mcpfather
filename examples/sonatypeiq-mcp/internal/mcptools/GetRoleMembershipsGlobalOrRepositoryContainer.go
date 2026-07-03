package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetRoleMembershipsGlobalOrRepositoryContainer tool
const GetRoleMembershipsGlobalOrRepositoryContainerInputSchema = "{\n  \"properties\": {\n    \"ownerType\": {\n      \"description\": \"Enter the value for ownerType. Using " + "\x60" + "global" + "\x60" + " will return the users and groups who have been assigned the administrator role.\",\n      \"enum\": [\n        \"repository_container\",\n        \"global\"\n      ],\n      \"pattern\": \"global|repository_container\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetRoleMembershipsGlobalOrRepositoryContainer tool (Status: 200, Content-Type: application/json)
const GetRoleMembershipsGlobalOrRepositoryContainerResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains all role Ids and the corresponding users/user groups assigned to them, for the ownerType specified. It also includes members who inherit a role based on the organization hierarchy.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **memberMappings** (Type: array):\n    - **Items** (Type: object):\n      - **roleId** (Type: string):\n      - **members** (Type: array):\n        - **Items** (Type: object):\n          - **ownerId** (Type: string):\n          - **ownerType** (Type: string):\n          - **type** (Type: string):\n              - Enum: ['USER', 'GROUP']\n          - **userOrGroupName** (Type: string):\n"

// NewGetRoleMembershipsGlobalOrRepositoryContainerMCPTool creates the MCP Tool instance for GetRoleMembershipsGlobalOrRepositoryContainer
func NewGetRoleMembershipsGlobalOrRepositoryContainerMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRoleMembershipsGlobalOrRepositoryContainer",
		"Use this method to retrieve all users and roles globally or for all repositories.\n\nPermissions required: Edit system configuration and users for a global context or view IQ elements for a non-global context",
		[]byte(GetRoleMembershipsGlobalOrRepositoryContainerInputSchema),
	)
}

// GetRoleMembershipsGlobalOrRepositoryContainerHandler is the handler function for the GetRoleMembershipsGlobalOrRepositoryContainer tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRoleMembershipsGlobalOrRepositoryContainerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/roleMemberships/{ownerType}", args, []string{"ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetRoleMembershipsGlobalOrRepositoryContainer")
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
