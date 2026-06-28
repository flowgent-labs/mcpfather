package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the RevokeRoleMembershipGlobalOrRepositoryContainer tool
const RevokeRoleMembershipGlobalOrRepositoryContainerInputSchema = "{\n  \"properties\": {\n    \"memberName\": {\n      \"description\": \"Enter the value for memberName. This can be a username or group name depending upon the value of memberType above.\",\n      \"type\": \"string\"\n    },\n    \"memberType\": {\n      \"description\": \"Enter the value for memberType, to specify a user or a user group.\",\n      \"enum\": [\n        \"USER\",\n        \"GROUP\"\n      ],\n      \"pattern\": \"(?i:user|group)\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the value for ownerType. Using " + "\x60" + "global" + "\x60" + " will revoke the administrator role.\",\n      \"enum\": [\n        \"repository_container\",\n        \"global\"\n      ],\n      \"pattern\": \"global|repository_container\",\n      \"type\": \"string\"\n    },\n    \"roleId\": {\n      \"description\": \"Enter the roleId for the role to be revoked.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"memberName\",\n    \"memberType\",\n    \"ownerType\",\n    \"roleId\"\n  ],\n  \"type\": \"object\"\n}"

// NewRevokeRoleMembershipGlobalOrRepositoryContainerMCPTool creates the MCP Tool instance for RevokeRoleMembershipGlobalOrRepositoryContainer
func NewRevokeRoleMembershipGlobalOrRepositoryContainerMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"RevokeRoleMembershipGlobalOrRepositoryContainer",
		"Use this method to revoke roles globally or on all repositories.\n\nPermissions required: Edit system configuration and users for a global context or view IQ elements for a non-global context",
		[]byte(RevokeRoleMembershipGlobalOrRepositoryContainerInputSchema),
	)
}

// RevokeRoleMembershipGlobalOrRepositoryContainerHandler is the handler function for the RevokeRoleMembershipGlobalOrRepositoryContainer tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func RevokeRoleMembershipGlobalOrRepositoryContainerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "DELETE", "/api/v2/roleMemberships/{ownerType}/role/{roleId}/{memberType}/{memberName}", args, []string{"memberName", "memberType", "ownerType", "roleId"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "RevokeRoleMembershipGlobalOrRepositoryContainer")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "DELETE", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
