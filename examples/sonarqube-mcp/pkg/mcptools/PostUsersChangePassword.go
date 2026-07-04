package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostUsersChangePassword tool
const PostUsersChangePasswordInputSchema = "{\n  \"properties\": {\n    \"login\": {\n      \"description\": \"User login\",\n      \"type\": \"string\"\n    },\n    \"password\": {\n      \"description\": \"The password needs to fulfill the following requirements: at least 12 characters and contain at least one uppercase character, one lowercase character, one digit and one special character.\",\n      \"minLength\": 12,\n      \"type\": \"string\"\n    },\n    \"previousPassword\": {\n      \"description\": \"Previous password. Required when changing one's own password.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"login\",\n    \"password\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostUsersChangePasswordMCPTool creates the MCP Tool instance for PostUsersChangePassword
func NewPostUsersChangePasswordMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostUsersChangePassword",
		"Update a user's password. - Update a user's password. Authenticated users can change their own password, provided that the account is not linked to an external authentication system. Administer System permission is required to change another user's password.",
		[]byte(PostUsersChangePasswordInputSchema),
	)
}

// PostUsersChangePasswordHandler is the handler function for the PostUsersChangePassword tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostUsersChangePasswordHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/users/change_password", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostUsersChangePassword")
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
