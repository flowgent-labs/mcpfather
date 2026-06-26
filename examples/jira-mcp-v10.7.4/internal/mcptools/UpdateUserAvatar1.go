package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateUserAvatar1 tool
const UpdateUserAvatar1InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"New avatar details\",\n      \"properties\": {\n        \"id\": {\n          \"example\": \"1000\",\n          \"type\": \"string\"\n        },\n        \"owner\": {\n          \"example\": \"fred\",\n          \"type\": \"string\"\n        },\n        \"selected\": {\n          \"type\": \"boolean\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"username\": {\n      \"description\": \"username\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateUserAvatar1 tool (Status: 200, Content-Type: application/json)
const UpdateUserAvatar1ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns updated avatar\n\n## Response Structure\n\n- Structure (Type: object):\n  - **owner** (Type: string):\n      - Example: 'fred'\n  - **selected** (Type: boolean):\n  - **id** (Type: string):\n      - Example: '1000'\n"

// NewUpdateUserAvatar1MCPTool creates the MCP Tool instance for UpdateUserAvatar1
func NewUpdateUserAvatar1MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateUserAvatar1",
		"Update user avatar - Updates the avatar for the user.",
		[]byte(UpdateUserAvatar1InputSchema),
	)
}

// UpdateUserAvatar1Handler is the handler function for the UpdateUserAvatar1 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateUserAvatar1Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/user/avatar", args, []string{}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateUserAvatar1"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
