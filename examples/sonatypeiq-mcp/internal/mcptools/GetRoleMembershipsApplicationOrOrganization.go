package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetRoleMembershipsApplicationOrOrganization tool
const GetRoleMembershipsApplicationOrOrganizationInputSchema = "{\n  \"properties\": {\n    \"internalOwnerId\": {\n      \"description\": \"Enter the corresponding id for the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType for which you want to retrieve users and their role Ids.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"internalOwnerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the GetRoleMembershipsApplicationOrOrganization tool (Status: 200, Content-Type: application/json)
const GetRoleMembershipsApplicationOrOrganizationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains the assigned role Ids, users and user groups for the application or organization requested. It also includes members who inherit a role based on the organization hierarchy.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **memberMappings** (Type: array):\n    - **Items** (Type: object):\n      - **members** (Type: array):\n        - **Items** (Type: object):\n          - **ownerId** (Type: string):\n          - **ownerType** (Type: string):\n          - **type** (Type: string):\n              - Enum: ['USER', 'GROUP']\n          - **userOrGroupName** (Type: string):\n      - **roleId** (Type: string):\n"

// NewGetRoleMembershipsApplicationOrOrganizationMCPTool creates the MCP Tool instance for GetRoleMembershipsApplicationOrOrganization
func NewGetRoleMembershipsApplicationOrOrganizationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetRoleMembershipsApplicationOrOrganization",
		"Use this method to retrieve the users, user groups and the corresponding role Ids.\n\nPermissions required: Edit system configuration and users for a global context or view IQ elements for a non-global context",
		[]byte(GetRoleMembershipsApplicationOrOrganizationInputSchema),
	)
}

// GetRoleMembershipsApplicationOrOrganizationHandler is the handler function for the GetRoleMembershipsApplicationOrOrganization tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetRoleMembershipsApplicationOrOrganizationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/roleMemberships/{ownerType}/{internalOwnerId}", args, []string{"internalOwnerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetRoleMembershipsApplicationOrOrganization")
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
