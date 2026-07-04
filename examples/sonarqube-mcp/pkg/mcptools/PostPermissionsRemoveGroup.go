package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostPermissionsRemoveGroup tool
const PostPermissionsRemoveGroupInputSchema = "{\n  \"properties\": {\n    \"groupName\": {\n      \"description\": \"Group name or 'anyone' (case insensitive)\",\n      \"type\": \"string\"\n    },\n    \"permission\": {\n      \"description\": \"The permission you would like to revoke from the group. - Possible values for global permissions: admin, gateadmin, profileadmin, provisioning, scan, applicationcreator, portfoliocreator - Possible values for project permissions admin, codeviewer, issueadmin, securityhotspotadmin, architectureadmin, scan, user\",\n      \"type\": \"string\"\n    },\n    \"projectId\": {\n      \"description\": \"Project id\",\n      \"type\": \"string\"\n    },\n    \"projectKey\": {\n      \"description\": \"Project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"groupName\",\n    \"permission\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostPermissionsRemoveGroupMCPTool creates the MCP Tool instance for PostPermissionsRemoveGroup
func NewPostPermissionsRemoveGroupMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostPermissionsRemoveGroup",
		"Remove a permission from a group. - Remove a permission from a group. This service defaults to global permissions, but can be limited to project permissions by providing project id or project key. The group name must be provided. Requires one of the following permissions: - 'Administer System' - 'Administer' rights on the specified project",
		[]byte(PostPermissionsRemoveGroupInputSchema),
	)
}

// PostPermissionsRemoveGroupHandler is the handler function for the PostPermissionsRemoveGroup tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostPermissionsRemoveGroupHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/permissions/remove_group", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostPermissionsRemoveGroup")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
