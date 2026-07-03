package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the SetBulkRoleMembersGlobalOrRepositoryContainer tool
const SetBulkRoleMembersGlobalOrRepositoryContainerInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"List of members to assign to this role. Provide an empty list to remove all members.\",\n      \"items\": {\n        \"properties\": {\n          \"displayName\": {\n            \"type\": \"string\"\n          },\n          \"email\": {\n            \"type\": \"string\"\n          },\n          \"internalName\": {\n            \"type\": \"string\"\n          },\n          \"realm\": {\n            \"type\": \"string\"\n          },\n          \"type\": {\n            \"enum\": [\n              \"USER\",\n              \"GROUP\"\n            ],\n            \"type\": \"string\"\n          }\n        },\n        \"type\": \"object\"\n      },\n      \"type\": \"array\"\n    },\n    \"ownerType\": {\n      \"description\": \"Enter the ownerType.\",\n      \"enum\": [\n        \"repository_container\",\n        \"global\"\n      ],\n      \"pattern\": \"global|repository_container\",\n      \"type\": \"string\"\n    },\n    \"roleId\": {\n      \"description\": \"Enter the roleId for the role whose members should be set.\\n\\nUse the Roles REST API for roleIds and descriptions.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"ownerType\",\n    \"roleId\"\n  ],\n  \"type\": \"object\"\n}"

// NewSetBulkRoleMembersGlobalOrRepositoryContainerMCPTool creates the MCP Tool instance for SetBulkRoleMembersGlobalOrRepositoryContainer
func NewSetBulkRoleMembersGlobalOrRepositoryContainerMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetBulkRoleMembersGlobalOrRepositoryContainer",
		"Use this method to set all members for a specific role in the global or repository container context. This operation atomically replaces all existing members for the role with the provided list.\n\nPermissions required: Edit system configuration and users for a global context or edit access control for a non-global context",
		[]byte(SetBulkRoleMembersGlobalOrRepositoryContainerInputSchema),
	)
}

// SetBulkRoleMembersGlobalOrRepositoryContainerHandler is the handler function for the SetBulkRoleMembersGlobalOrRepositoryContainer tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetBulkRoleMembersGlobalOrRepositoryContainerHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/roleMemberships/{ownerType}/role/{roleId}/members", args, []string{"ownerType", "roleId"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetBulkRoleMembersGlobalOrRepositoryContainer")
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
