package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostUserGroupsAddUser tool
const PostUserGroupsAddUserInputSchema = "{\n  \"properties\": {\n    \"login\": {\n      \"description\": \"User login\",\n      \"type\": \"string\"\n    },\n    \"name\": {\n      \"description\": \"Group name\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"name\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostUserGroupsAddUserMCPTool creates the MCP Tool instance for PostUserGroupsAddUser
func NewPostUserGroupsAddUserMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostUserGroupsAddUser",
		"Add a user to a group. - Add a user to a group. 'name' must be provided. Requires the following permission: 'Administer System'.",
		[]byte(PostUserGroupsAddUserInputSchema),
	)
}

// PostUserGroupsAddUserHandler is the handler function for the PostUserGroupsAddUser tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostUserGroupsAddUserHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/user_groups/add_user", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostUserGroupsAddUser")
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
