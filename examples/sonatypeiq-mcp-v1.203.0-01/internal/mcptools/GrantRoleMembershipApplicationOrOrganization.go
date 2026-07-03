package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GrantRoleMembershipApplicationOrOrganization tool
const GrantRoleMembershipApplicationOrOrganizationInputSchema = "{\n  \"properties\": {\n    \"internalOwnerId\": {\n      \"description\": \"Enter the value for the internalId associated with the ownerType specified above.\",\n      \"type\": \"string\"\n    },\n    \"memberName\": {\n      \"description\": \"Enter the value for memberName. This can be a username or group name depending upon the value of memberType above.\",\n      \"type\": \"string\"\n    },\n    \"memberType\": {\n      \"description\": \"Enter the value for memberType, to specify a user or a user group.\",\n      \"enum\": [\n        \"USER\",\n        \"GROUP\"\n      ],\n      \"pattern\": \"(?i:user|group)\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the value for the ownerType for which you want to grant the role.\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    },\n    \"roleId\": {\n      \"description\": \"Enter the roleId for the role to be granted.\\n\\nUse the Roles REST API for roleIds and descriptions.\",\n      \"type\": \"string\"\n    },\n    \"validateMember\": {\n      \"default\": false,\n      \"description\": \"If true, attempts to validate if the specified user or group exists before assigning the role.\\nIf false or omitted, the request behaves as before (no validation check).\",\n      \"type\": \"boolean\"\n    }\n  },\n  \"required\": [\n    \"internalOwnerId\",\n    \"memberName\",\n    \"memberType\",\n    \"ownerType\",\n    \"roleId\"\n  ],\n  \"type\": \"object\"\n}"

// NewGrantRoleMembershipApplicationOrOrganizationMCPTool creates the MCP Tool instance for GrantRoleMembershipApplicationOrOrganization
func NewGrantRoleMembershipApplicationOrOrganizationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GrantRoleMembershipApplicationOrOrganization",
		"Use this method to grant a role to a user or user group for the specified application or organization.\n\nPermissions required: Edit access control",
		[]byte(GrantRoleMembershipApplicationOrOrganizationInputSchema),
	)
}

// GrantRoleMembershipApplicationOrOrganizationHandler is the handler function for the GrantRoleMembershipApplicationOrOrganization tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GrantRoleMembershipApplicationOrOrganizationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/roleMemberships/{ownerType}/{internalOwnerId}/role/{roleId}/{memberType}/{memberName}", args, []string{"internalOwnerId", "memberName", "memberType", "ownerType", "roleId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GrantRoleMembershipApplicationOrOrganization")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
