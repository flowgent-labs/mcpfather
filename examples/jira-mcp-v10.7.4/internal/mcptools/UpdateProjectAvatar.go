package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"time"
)

// Input Schema for the UpdateProjectAvatar tool
const UpdateProjectAvatarInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Avatar data\",\n      \"properties\": {\n        \"id\": {\n          \"example\": \"1000\",\n          \"type\": \"string\"\n        },\n        \"owner\": {\n          \"example\": \"fred\",\n          \"type\": \"string\"\n        },\n        \"selected\": {\n          \"type\": \"boolean\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"projectIdOrKey\": {\n      \"description\": \"Project id or project key\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"body\",\n    \"projectIdOrKey\"\n  ],\n  \"type\": \"object\"\n}"

// NewUpdateProjectAvatarMCPTool creates the MCP Tool instance for UpdateProjectAvatar
func NewUpdateProjectAvatarMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateProjectAvatar",
		"Update project avatar - Updates an avatar for a project. This is step 3/3 of changing an avatar for a project.",
		[]byte(UpdateProjectAvatarInputSchema),
	)
}

// UpdateProjectAvatarHandler is the handler function for the UpdateProjectAvatar tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateProjectAvatarHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/rest/api/2/project/{projectIdOrKey}/avatar", args, []string{"projectIdOrKey"}, contentType)
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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateProjectAvatar"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
