package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the ChangeUserPassword tool
const ChangeUserPasswordInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Password details\",\n      \"properties\": {\n        \"currentPassword\": {\n          \"example\": \"current password\",\n          \"type\": \"string\"\n        },\n        \"password\": {\n          \"example\": \"new password\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"key\": {\n      \"description\": \"user key\",\n      \"type\": \"string\"\n    },\n    \"username\": {\n      \"description\": \"the username\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// NewChangeUserPasswordMCPTool creates the MCP Tool instance for ChangeUserPassword
func NewChangeUserPasswordMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ChangeUserPassword",
		"Update user password - Modify user password.",
		[]byte(ChangeUserPasswordInputSchema),
	)
}

// ChangeUserPasswordHandler is the handler function for the ChangeUserPassword tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ChangeUserPasswordHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/user/password", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "ChangeUserPassword"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
